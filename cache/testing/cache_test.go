package testing

import (
	"context"
	"testing"
	"time"

	"github.com/cjcox17/sudokit/cache"
)

func TestMockCacheService_SetGet(t *testing.T) {
	svc := NewMockCacheService()

	err := svc.Set(context.Background(), "key1", []byte("value1"), 5*time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	val, err := svc.Get(context.Background(), "key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(val) != "value1" {
		t.Errorf("expected value1, got %s", string(val))
	}
}

func TestMockCacheService_GetString(t *testing.T) {
	svc := NewMockCacheService()

	svc.SetString(context.Background(), "key1", "hello", 5*time.Minute)

	val, err := svc.GetString(context.Background(), "key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if val != "hello" {
		t.Errorf("expected hello, got %s", val)
	}
}

func TestMockCacheService_GetInt(t *testing.T) {
	svc := NewMockCacheService()

	svc.SetInt(context.Background(), "counter", 42, 5*time.Minute)

	val, err := svc.GetInt(context.Background(), "counter")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if val != 42 {
		t.Errorf("expected 42, got %d", val)
	}
}

func TestMockCacheService_GetJSON(t *testing.T) {
	svc := NewMockCacheService()

	type User struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}

	svc.SetJSON(context.Background(), "user:1", User{Name: "John", Age: 30}, 5*time.Minute)

	var user User
	err := svc.GetJSON(context.Background(), "user:1", &user)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if user.Name != "John" {
		t.Errorf("expected John, got %s", user.Name)
	}
}

func TestMockCacheService_Delete(t *testing.T) {
	svc := NewMockCacheService()

	svc.Set(context.Background(), "key1", []byte("value1"), 5*time.Minute)

	err := svc.Delete(context.Background(), "key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	_, err = svc.Get(context.Background(), "key1")
	if err != cache.ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestMockCacheService_Exists(t *testing.T) {
	svc := NewMockCacheService()

	svc.Set(context.Background(), "key1", []byte("value1"), 5*time.Minute)

	exists, err := svc.Exists(context.Background(), "key1")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if !exists {
		t.Error("expected key to exist")
	}

	exists, err = svc.Exists(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if exists {
		t.Error("expected key to not exist")
	}
}

func TestMockCacheService_IncrDecr(t *testing.T) {
	svc := NewMockCacheService()

	val, err := svc.Incr(context.Background(), "counter")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}

	val, err = svc.Incr(context.Background(), "counter")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if val != 2 {
		t.Errorf("expected 2, got %d", val)
	}

	val, err = svc.Decr(context.Background(), "counter")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if val != 1 {
		t.Errorf("expected 1, got %d", val)
	}
}

func TestMockCacheService_Remember(t *testing.T) {
	svc := NewMockCacheService()

	callCount := 0

	val, err := svc.Remember(context.Background(), "key1", 5*time.Minute, func() ([]byte, error) {
		callCount++
		return []byte("computed"), nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(val) != "computed" {
		t.Errorf("expected computed, got %s", string(val))
	}

	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}

	val, err = svc.Remember(context.Background(), "key1", 5*time.Minute, func() ([]byte, error) {
		callCount++
		return []byte("computed again"), nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(val) != "computed" {
		t.Errorf("expected computed (cached), got %s", string(val))
	}

	if callCount != 1 {
		t.Errorf("expected 1 call (cached), got %d", callCount)
	}
}

func TestMockCacheService_Assertions(t *testing.T) {
	svc := NewMockCacheService()

	svc.Set(context.Background(), "key1", []byte("value1"), 5*time.Minute)

	svc.AssertKeyExists(t, "key1")
	svc.AssertKeyValue(t, "key1", []byte("value1"))
	svc.AssertKeyNotExists(t, "key2")
}

func TestMockCacheService_Reset(t *testing.T) {
	svc := NewMockCacheService()

	svc.Set(context.Background(), "key1", []byte("value1"), 5*time.Minute)
	svc.Set(context.Background(), "key2", []byte("value2"), 5*time.Minute)

	if len(svc.GetItems()) != 2 {
		t.Error("expected 2 items before reset")
	}

	svc.Reset()

	if len(svc.GetItems()) != 0 {
		t.Error("expected 0 items after reset")
	}
}

func TestMockCacheService_KeyNotFound(t *testing.T) {
	svc := NewMockCacheService()

	_, err := svc.Get(context.Background(), "nonexistent")
	if err != cache.ErrKeyNotFound {
		t.Errorf("expected ErrKeyNotFound, got %v", err)
	}
}

func TestCacheProviderConstants(t *testing.T) {
	if cache.ProviderRedis != "redis" {
		t.Error("expected ProviderRedis to be 'redis'")
	}
	if cache.ProviderMemory != "memory" {
		t.Error("expected ProviderMemory to be 'memory'")
	}
	if cache.ProviderMock != "mock" {
		t.Error("expected ProviderMock to be 'mock'")
	}
}

func TestMockCacheService_CustomBehavior(t *testing.T) {
	svc := NewMockCacheService()

	svc.GetFunc = func(ctx context.Context, key string) ([]byte, error) {
		return []byte("custom-value"), nil
	}

	val, err := svc.Get(context.Background(), "any-key")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if string(val) != "custom-value" {
		t.Errorf("expected custom-value, got %s", string(val))
	}
}
