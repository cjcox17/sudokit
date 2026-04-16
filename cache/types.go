package cache

import (
	"context"
	"errors"
	"time"
)

var (
	ErrKeyNotFound       = errors.New("key not found")
	ErrInvalidKey        = errors.New("invalid key")
	ErrInvalidValue      = errors.New("invalid value")
	ErrInvalidTTL        = errors.New("invalid TTL")
	ErrSetFailed         = errors.New("failed to set value")
	ErrGetFailed         = errors.New("failed to get value")
	ErrDeleteFailed      = errors.New("failed to delete value")
	ErrProviderNotConfig = errors.New("cache provider not configured")
)

type Provider string

const (
	ProviderRedis  Provider = "redis"
	ProviderMemory Provider = "memory"
	ProviderMock   Provider = "mock"
)

type Item struct {
	Key       string
	Value     []byte
	ExpiresAt time.Time
}

type Config struct {
	Provider      Provider
	RedisAddr     string
	RedisPassword string
	RedisDB       int
	DefaultTTL    time.Duration
	MaxMemoryMB   int
}

type Store interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Keys(ctx context.Context, pattern string) ([]string, error)
	Flush(ctx context.Context) error
	HealthCheck(ctx context.Context) error
}

type Service interface {
	Get(ctx context.Context, key string) ([]byte, error)
	GetString(ctx context.Context, key string) (string, error)
	GetInt(ctx context.Context, key string) (int64, error)
	GetJSON(ctx context.Context, key string, v interface{}) error
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	SetString(ctx context.Context, key string, value string, ttl time.Duration) error
	SetInt(ctx context.Context, key string, value int64, ttl time.Duration) error
	SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	Expire(ctx context.Context, key string, ttl time.Duration) error
	TTL(ctx context.Context, key string) (time.Duration, error)
	Incr(ctx context.Context, key string) (int64, error)
	Decr(ctx context.Context, key string) (int64, error)
	Remember(ctx context.Context, key string, ttl time.Duration, fn func() ([]byte, error)) ([]byte, error)
	RememberString(ctx context.Context, key string, ttl time.Duration, fn func() (string, error)) (string, error)
	RememberJSON(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error), v interface{}) error
	HealthCheck(ctx context.Context) error
}
