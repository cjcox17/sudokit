package testing

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/cjcox17/sudokit/cache"
)

type MockCacheService struct {
	mu sync.RWMutex

	GetFunc         func(ctx context.Context, key string) ([]byte, error)
	SetFunc         func(ctx context.Context, key string, value []byte, ttl time.Duration) error
	DeleteFunc      func(ctx context.Context, key string) error
	ExistsFunc      func(ctx context.Context, key string) (bool, error)
	ExpireFunc      func(ctx context.Context, key string, ttl time.Duration) error
	TTLFunc         func(ctx context.Context, key string) (time.Duration, error)
	IncrFunc        func(ctx context.Context, key string) (int64, error)
	DecrFunc        func(ctx context.Context, key string) (int64, error)
	KeysFunc        func(ctx context.Context, pattern string) ([]string, error)
	FlushFunc       func(ctx context.Context) error
	HealthCheckFunc func(ctx context.Context) error

	GetCalls    []GetCall
	SetCalls    []SetCall
	DeleteCalls []DeleteCall
	ExistsCalls []ExistsCall
	IncrCalls   []IncrCall

	items map[string][]byte
}

type GetCall struct {
	Key string
}

type SetCall struct {
	Key   string
	Value []byte
	TTL   time.Duration
}

type DeleteCall struct {
	Key string
}

type ExistsCall struct {
	Key string
}

type IncrCall struct {
	Key string
}

func NewMockCacheService() *MockCacheService {
	return &MockCacheService{
		items: make(map[string][]byte),
	}
}

func (s *MockCacheService) Get(ctx context.Context, key string) ([]byte, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.GetCalls = append(s.GetCalls, GetCall{Key: key})

	if s.GetFunc != nil {
		return s.GetFunc(ctx, key)
	}

	data, ok := s.items[key]
	if !ok {
		return nil, cache.ErrKeyNotFound
	}
	return data, nil
}

func (s *MockCacheService) GetString(ctx context.Context, key string) (string, error) {
	data, err := s.Get(ctx, key)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func (s *MockCacheService) GetInt(ctx context.Context, key string) (int64, error) {
	data, err := s.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	var val int64
	json.Unmarshal(data, &val)
	return val, nil
}

func (s *MockCacheService) GetJSON(ctx context.Context, key string, v interface{}) error {
	data, err := s.Get(ctx, key)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

func (s *MockCacheService) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SetCalls = append(s.SetCalls, SetCall{Key: key, Value: value, TTL: ttl})

	if s.SetFunc != nil {
		return s.SetFunc(ctx, key, value, ttl)
	}

	s.items[key] = value
	return nil
}

func (s *MockCacheService) SetString(ctx context.Context, key string, value string, ttl time.Duration) error {
	return s.Set(ctx, key, []byte(value), ttl)
}

func (s *MockCacheService) SetInt(ctx context.Context, key string, value int64, ttl time.Duration) error {
	data, _ := json.Marshal(value)
	return s.Set(ctx, key, data, ttl)
}

func (s *MockCacheService) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.Set(ctx, key, data, ttl)
}

func (s *MockCacheService) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.DeleteCalls = append(s.DeleteCalls, DeleteCall{Key: key})

	if s.DeleteFunc != nil {
		return s.DeleteFunc(ctx, key)
	}

	delete(s.items, key)
	return nil
}

func (s *MockCacheService) DeleteByPattern(ctx context.Context, pattern string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for key := range s.items {
		if matchPattern(key, pattern) {
			delete(s.items, key)
		}
	}
	return nil
}

func (s *MockCacheService) Exists(ctx context.Context, key string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.ExistsCalls = append(s.ExistsCalls, ExistsCall{Key: key})

	if s.ExistsFunc != nil {
		return s.ExistsFunc(ctx, key)
	}

	_, ok := s.items[key]
	return ok, nil
}

func (s *MockCacheService) Expire(ctx context.Context, key string, ttl time.Duration) error {
	if s.ExpireFunc != nil {
		return s.ExpireFunc(ctx, key, ttl)
	}
	return nil
}

func (s *MockCacheService) TTL(ctx context.Context, key string) (time.Duration, error) {
	if s.TTLFunc != nil {
		return s.TTLFunc(ctx, key)
	}
	return 0, nil
}

func (s *MockCacheService) Incr(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.IncrCalls = append(s.IncrCalls, IncrCall{Key: key})

	if s.IncrFunc != nil {
		return s.IncrFunc(ctx, key)
	}

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

func (s *MockCacheService) Decr(ctx context.Context, key string) (int64, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.DecrFunc != nil {
		return s.DecrFunc(ctx, key)
	}

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

func (s *MockCacheService) Remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error) {
	data, err := s.Get(ctx, key)
	if err == nil {
		return data, nil
	}

	data, err = fn()
	if err != nil {
		return nil, err
	}

	s.Set(ctx, key, data, ttl)
	return data, nil
}

func (s *MockCacheService) RememberString(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error) {
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

func (s *MockCacheService) RememberJSON(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), v interface{}) error {
	data, err := s.Get(ctx, key)
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

	s.Set(ctx, key, data, ttl)
	return json.Unmarshal(data, v)
}

func (s *MockCacheService) HealthCheck(ctx context.Context) error {
	if s.HealthCheckFunc != nil {
		return s.HealthCheckFunc(ctx)
	}
	return nil
}

func (s *MockCacheService) GetItems() map[string][]byte {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string][]byte)
	for k, v := range s.items {
		result[k] = v
	}
	return result
}

func (s *MockCacheService) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.items = make(map[string][]byte)
	s.GetCalls = nil
	s.SetCalls = nil
	s.DeleteCalls = nil
	s.ExistsCalls = nil
	s.IncrCalls = nil
}

func (s *MockCacheService) AssertKeyExists(t assertT, key string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.items[key]; !ok {
		t.Errorf("expected key %s to exist in cache", key)
	}
}

func (s *MockCacheService) AssertKeyNotExists(t assertT, key string) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.items[key]; ok {
		t.Errorf("expected key %s to not exist in cache", key)
	}
}

func (s *MockCacheService) AssertKeyValue(t assertT, key string, expected []byte) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	actual, ok := s.items[key]
	if !ok {
		t.Errorf("expected key %s to exist in cache", key)
		return
	}

	if string(actual) != string(expected) {
		t.Errorf("expected key %s to have value %s, got %s", key, string(expected), string(actual))
	}
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

type assertT interface {
	Errorf(format string, args ...interface{})
}
