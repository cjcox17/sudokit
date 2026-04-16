package testing

import (
	"bytes"
	"context"
	"io"
	"sync"
	"time"

	"github.com/cjcox17/sudokit/storage"
)

type MockStorageService struct {
	mu sync.RWMutex

	UploadFunc        func(ctx context.Context, key string, reader io.Reader, contentType string) error
	UploadBytesFunc   func(ctx context.Context, key string, data []byte, contentType string) error
	DownloadFunc      func(ctx context.Context, key string) (io.ReadCloser, error)
	DownloadBytesFunc func(ctx context.Context, key string) ([]byte, error)
	DeleteFunc        func(ctx context.Context, key string) error
	ExistsFunc        func(ctx context.Context, key string) (bool, error)
	PresignedURLFunc  func(ctx context.Context, key string, expiry time.Duration) (string, error)

	UploadCalls       []UploadCall
	DownloadCalls     []DownloadCall
	DeleteCalls       []DeleteCall
	ExistsCalls       []ExistsCall
	PresignedURLCalls []PresignedURLCall

	files map[string][]byte
}

type UploadCall struct {
	Key         string
	ContentType string
}

type DownloadCall struct {
	Key string
}

type DeleteCall struct {
	Key string
}

type ExistsCall struct {
	Key string
}

type PresignedURLCall struct {
	Key    string
	Expiry time.Duration
}

func NewMockStorageService() *MockStorageService {
	return &MockStorageService{
		UploadCalls:       make([]UploadCall, 0),
		DownloadCalls:     make([]DownloadCall, 0),
		DeleteCalls:       make([]DeleteCall, 0),
		ExistsCalls:       make([]ExistsCall, 0),
		PresignedURLCalls: make([]PresignedURLCall, 0),
		files:             make(map[string][]byte),
	}
}

func (m *MockStorageService) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	m.mu.Lock()
	m.UploadCalls = append(m.UploadCalls, UploadCall{
		Key:         key,
		ContentType: contentType,
	})
	m.mu.Unlock()

	if m.UploadFunc != nil {
		return m.UploadFunc(ctx, key, reader, contentType)
	}

	data, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	m.mu.Lock()
	m.files[key] = data
	m.mu.Unlock()

	return nil
}

func (m *MockStorageService) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	m.mu.Lock()
	m.UploadCalls = append(m.UploadCalls, UploadCall{
		Key:         key,
		ContentType: contentType,
	})
	m.mu.Unlock()

	if m.UploadBytesFunc != nil {
		return m.UploadBytesFunc(ctx, key, data, contentType)
	}

	m.mu.Lock()
	m.files[key] = data
	m.mu.Unlock()

	return nil
}

func (m *MockStorageService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	m.mu.Lock()
	m.DownloadCalls = append(m.DownloadCalls, DownloadCall{Key: key})
	m.mu.Unlock()

	if m.DownloadFunc != nil {
		return m.DownloadFunc(ctx, key)
	}

	m.mu.RLock()
	data, ok := m.files[key]
	m.mu.RUnlock()

	if !ok {
		return nil, storage.ErrNotFound
	}

	return io.NopCloser(bytes.NewReader(data)), nil
}

func (m *MockStorageService) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	m.DownloadCalls = append(m.DownloadCalls, DownloadCall{Key: key})
	m.mu.Unlock()

	if m.DownloadBytesFunc != nil {
		return m.DownloadBytesFunc(ctx, key)
	}

	m.mu.RLock()
	data, ok := m.files[key]
	m.mu.RUnlock()

	if !ok {
		return nil, storage.ErrNotFound
	}

	return data, nil
}

func (m *MockStorageService) Delete(ctx context.Context, key string) error {
	m.mu.Lock()
	m.DeleteCalls = append(m.DeleteCalls, DeleteCall{Key: key})
	delete(m.files, key)
	m.mu.Unlock()

	if m.DeleteFunc != nil {
		return m.DeleteFunc(ctx, key)
	}

	return nil
}

func (m *MockStorageService) DeleteMultiple(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := m.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (m *MockStorageService) Exists(ctx context.Context, key string) (bool, error) {
	m.mu.Lock()
	m.ExistsCalls = append(m.ExistsCalls, ExistsCall{Key: key})
	m.mu.Unlock()

	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, key)
	}

	m.mu.RLock()
	_, ok := m.files[key]
	m.mu.RUnlock()

	return ok, nil
}

func (m *MockStorageService) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	m.mu.Lock()
	m.PresignedURLCalls = append(m.PresignedURLCalls, PresignedURLCall{
		Key:    key,
		Expiry: expiry,
	})
	m.mu.Unlock()

	if m.PresignedURLFunc != nil {
		return m.PresignedURLFunc(ctx, key, expiry)
	}

	return "https://mock-presigned-url.com/" + key, nil
}

func (m *MockStorageService) PresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	return m.PresignedURL(ctx, key, expiry)
}

func (m *MockStorageService) Copy(ctx context.Context, srcKey, dstKey string) error {
	m.mu.RLock()
	data, ok := m.files[srcKey]
	m.mu.RUnlock()

	if !ok {
		return storage.ErrNotFound
	}

	m.mu.Lock()
	m.files[dstKey] = data
	m.mu.Unlock()

	return nil
}

func (m *MockStorageService) GetMetadata(ctx context.Context, key string) (map[string]string, error) {
	m.mu.RLock()
	_, ok := m.files[key]
	m.mu.RUnlock()

	if !ok {
		return nil, storage.ErrNotFound
	}

	return map[string]string{
		"content-type": "application/octet-stream",
	}, nil
}

func (m *MockStorageService) List(ctx context.Context, prefix string) ([]string, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var keys []string
	for key := range m.files {
		if prefix == "" || len(key) >= len(prefix) && key[:len(prefix)] == prefix {
			keys = append(keys, key)
		}
	}
	return keys, nil
}

func (m *MockStorageService) Reset() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.UploadCalls = make([]UploadCall, 0)
	m.DownloadCalls = make([]DownloadCall, 0)
	m.DeleteCalls = make([]DeleteCall, 0)
	m.ExistsCalls = make([]ExistsCall, 0)
	m.PresignedURLCalls = make([]PresignedURLCall, 0)
	m.files = make(map[string][]byte)
}

func (m *MockStorageService) SetFile(key string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[key] = data
}

func (m *MockStorageService) GetFile(key string) ([]byte, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	data, ok := m.files[key]
	return data, ok
}

func AssertFileUploaded(t interface{ Errorf(string, ...any) }, mock *MockStorageService, expectedKey string) {
	found := false
	for _, call := range mock.UploadCalls {
		if call.Key == expectedKey {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected file '%s' to be uploaded, but it was not", expectedKey)
	}
}

func AssertFileDeleted(t interface{ Errorf(string, ...any) }, mock *MockStorageService, expectedKey string) {
	found := false
	for _, call := range mock.DeleteCalls {
		if call.Key == expectedKey {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("expected file '%s' to be deleted, but it was not", expectedKey)
	}
}
