package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

type CacheService struct {
	store      Store
	defaultTTL time.Duration
}

func NewService(cfg *Config) (*CacheService, error) {
	var store Store

	switch cfg.Provider {
	case ProviderRedis:
		if cfg.RedisAddr == "" {
			return nil, ErrProviderNotConfig
		}
		store = NewRedisStore(cfg.RedisAddr, cfg.RedisPassword, cfg.RedisDB)
	case ProviderMemory:
		store = NewMemoryStore()
	case ProviderMock:
		store = NewMockStore()
	default:
		return nil, ErrProviderNotConfig
	}

	defaultTTL := cfg.DefaultTTL
	if defaultTTL == 0 {
		defaultTTL = 15 * time.Minute
	}

	return &CacheService{
		store:      store,
		defaultTTL: defaultTTL,
	}, nil
}

func (s *CacheService) Get(ctx context.Context, key string) ([]byte, error) {
	return s.store.Get(ctx, key)
}

func (s *CacheService) GetString(ctx context.Context, key string) (string, error) {
	data, err := s.store.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *CacheService) GetInt(ctx context.Context, key string) (int64, error) {
	data, err := s.store.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	var val int64
	if err := json.Unmarshal(data, &val); err != nil {
		return 0, fmt.Errorf("%w: %v", ErrInvalidValue, err)
	}
	return val, nil
}

func (s *CacheService) GetJSON(ctx context.Context, key string, v interface{}) error {
	data, err := s.store.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *CacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		ttl = s.defaultTTL
	}
	return s.store.Set(ctx, key, value, ttl)
}

func (s *CacheService) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.Set(ctx, key, []byte(value), ttl)
}

func (s *CacheService) SetInt(ctx context.Context, key string, value int64, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSetFailed, err)
	}
	return s.Set(ctx, key, data, ttl)
}

func (s *CacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSetFailed, err)
	}
	return s.Set(ctx, key, data, ttl)
}

func (s *CacheService) Delete(ctx context.Context, key string) error {
	return s.store.Delete(ctx, key)
}

func (s *CacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	keys, err := s.store.Keys(ctx, pattern)
	if err != nil {
		return err
	}
	for _, key := range keys {
		if err := s.store.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *CacheService) Exists(ctx context.Context, key string) (bool, error) {
	return s.store.Exists(ctx, key)
}

func (s *CacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return s.store.Expire(ctx, key, ttl)
}

func (s *CacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	return s.store.TTL(ctx, key)
}

func (s *CacheService) Incr(ctx context.Context, key string) (int64, error) {
	return s.store.Incr(ctx, key)
}

func (s *CacheService) Decr(ctx context.Context, key string) (int64, error) {
	return s.store.Decr(ctx, key)
}

func (s *CacheService) Remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error) {
	data, err := s.store.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	data, err = fn()
	if err != nil {
		return nil, err
	}

	if err := s.Set(ctx, key, data, ttl); err != nil {
		return nil, err
	}

	return data, nil
}

func (s *CacheService) RememberString(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
	data, err := s.Remember(ctx, key, ttl, func() ([]byte, error) {
		s, err := fn()
		if err != nil {
			return nil, err
		}
		return []byte(s), nil
	})
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *CacheService) RememberJSON(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), v interface{}) error {
	data, err := s.store.Get(ctx, key)
	if err == nil {
		return json.Unmarshal(data, v)
	}

	val, err := fn()
	if err != nil {
		return err
	}

	data, err = json.Marshal(val)
	if err != nil {
		return err
	}

	if err := s.Set(ctx, key, data, ttl); err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

func (s *CacheService) HealthCheck(ctx context.Context) error {
	return s.store.HealthCheck(ctx)
}

type MemoryStore struct {
	items map[string]*memoryItem
	mu    sync.RWMutex
}

type memoryItem struct {
	value     []byte
	expiresAt time.Time
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items: make(map[string]*memoryItem),
	}
}

func (s *MemoryStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return nil, ErrKeyNotFound
	}

	return item.value, nil
}

func (s *MemoryStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var expiresAt time.Time
	if ttl > 0 {
		expiresAt = time.Now().Add(ttl)
	}

	s.items[key] = &memoryItem{
		value:     value,
		expiresAt: expiresAt,
	}

	return nil
}

func (s *MemoryStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
	return nil
}

func (s *MemoryStore) Exists(ctx context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[key]
	if !ok {
		return false, nil
	}

	if !item.expiresAt.IsZero() && time.Now().After(item.expiresAt) {
		return false, nil
	}

	return true, nil
}

func (s *MemoryStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[key]
	if !ok {
		return ErrKeyNotFound
	}

	item.expiresAt = time.Now().Add(ttl)
	return nil
}

func (s *MemoryStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[key]
	if !ok {
		return 0, ErrKeyNotFound
	}

	if item.expiresAt.IsZero() {
		return -1, nil
	}

	ttl := time.Until(item.expiresAt)
	if ttl < 0 {
		return 0, nil
	}

	return ttl, nil
}

func (s *MemoryStore) Incr(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[key]
	if !ok {
		s.items[key] = &memoryItem{value: []byte("1")}
		return 1, nil
	}

	var val int64
	if err := json.Unmarshal(item.value, &val); err != nil {
		val = 0
	}
	val++
	data, _ := json.Marshal(val)
	item.value = data

	return val, nil
}

func (s *MemoryStore) Decr(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[key]
	if !ok {
		s.items[key] = &memoryItem{value: []byte("-1")}
		return -1, nil
	}

	var val int64
	if err := json.Unmarshal(item.value, &val); err != nil {
		val = 0
	}
	val--
	data, _ := json.Marshal(val)
	item.value = data

	return val, nil
}

func (s *MemoryStore) Keys(ctx context.Context, pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0)
	for key := range s.items {
		if matchPattern(key, pattern) {
			keys = append(keys, key)
		}
	}

	return keys, nil
}

func (s *MemoryStore) Flush(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string]*memoryItem)
	return nil
}

func (s *MemoryStore) HealthCheck(ctx context.Context) error {
	return nil
}

func matchPattern(key, pattern string) bool {
	if pattern == "*" {
		return true
	}
	if pattern == key {
		return true
	}
	return false
}

type MockStore struct {
	items map[string][]byte
	mu    sync.RWMutex
}

func NewMockStore() *MockStore {
	return &MockStore{
		items: make(map[string][]byte),
	}
}

func (s *MockStore) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	data, ok := s.items[key]
	if !ok {
		return nil, ErrKeyNotFound
	}

	return data, nil
}

func (s *MockStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items[key] = value
	return nil
}

func (s *MockStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.items, key)
	return nil
}

func (s *MockStore) Exists(ctx context.Context, key string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, ok := s.items[key]
	return ok, nil
}

func (s *MockStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return nil
}

func (s *MockStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (s *MockStore) Incr(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.items[key]
	var val int64
	if ok {
		json.Unmarshal(data, &val)
	}
	val++
	data, _ = json.Marshal(val)
	s.items[key] = data
	return val, nil
}

func (s *MockStore) Decr(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, ok := s.items[key]
	var val int64
	if ok {
		json.Unmarshal(data, &val)
	}
	val--
	data, _ = json.Marshal(val)
	s.items[key] = data
	return val, nil
}

func (s *MockStore) Keys(ctx context.Context, pattern string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	keys := make([]string, 0, len(s.items))
	for key := range s.items {
		keys = append(keys, key)
	}
	return keys, nil
}

func (s *MockStore) Flush(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string][]byte)
	return nil
}

func (s *MockStore) HealthCheck(ctx context.Context) error {
	return nil
}

func (s *MockStore) GetItems() map[string][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]byte)
	for k, v := range s.items {
		result[k] = v
	}
	return result
}

func (s *MockStore) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string][]byte)
}
