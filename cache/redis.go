package cache

import (
	"context"
	"time"
)

type RedisStore struct {
	addr     string
	password string
	db       int
}

func NewRedisStore(addr, password string, db int) *RedisStore {
	return &RedisStore{
		addr:     addr,
		password: password,
		db:       db,
	}
}

func (s *RedisStore) Get(ctx context.Context, key string) ([]byte, error) {
	return nil, ErrProviderNotConfig
}

func (s *RedisStore) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return ErrProviderNotConfig
}

func (s *RedisStore) Delete(ctx context.Context, key string) error {
	return ErrProviderNotConfig
}

func (s *RedisStore) Exists(ctx context.Context, key string) (bool, error) {
	return false, ErrProviderNotConfig
}

func (s *RedisStore) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return ErrProviderNotConfig
}

func (s *RedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, ErrProviderNotConfig
}

func (s *RedisStore) Incr(ctx context.Context, key string) (int64, error) {
	return 0, ErrProviderNotConfig
}

func (s *RedisStore) Decr(ctx context.Context, key string) (int64, error) {
	return 0, ErrProviderNotConfig
}

func (s *RedisStore) Keys(ctx context.Context, pattern string) ([]string, error) {
	return nil, ErrProviderNotConfig
}

func (s *RedisStore) Flush(ctx context.Context) error {
	return ErrProviderNotConfig
}

func (s *RedisStore) HealthCheck(ctx context.Context) error {
	return ErrProviderNotConfig
}
