package testing

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/cjcox17/sudokit/storage"
)

func TestMockStorageService_Upload(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	data := []byte("test content")
	err := mock.UploadBytes(ctx, "test.txt", data, "text/plain")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.UploadCalls) != 1 {
		t.Errorf("expected 1 upload call, got %d", len(mock.UploadCalls))
	}

	call := mock.UploadCalls[0]
	if call.Key != "test.txt" {
		t.Errorf("expected key 'test.txt', got '%s'", call.Key)
	}

	if call.ContentType != "text/plain" {
		t.Errorf("expected content type 'text/plain', got '%s'", call.ContentType)
	}
}

func TestMockStorageService_UploadReader(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	reader := bytes.NewReader([]byte("test content"))
	err := mock.Upload(ctx, "test.txt", reader, "text/plain")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.UploadCalls) != 1 {
		t.Errorf("expected 1 upload call, got %d", len(mock.UploadCalls))
	}
}

func TestMockStorageService_Download(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("test.txt", []byte("test content"))

	reader, err := mock.Download(ctx, "test.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Errorf("unexpected error reading: %v", err)
	}

	if string(data) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(data))
	}
}

func TestMockStorageService_DownloadBytes(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("test.txt", []byte("test content"))

	data, err := mock.DownloadBytes(ctx, "test.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(data) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(data))
	}
}

func TestMockStorageService_DownloadNotFound(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	_, err := mock.DownloadBytes(ctx, "nonexistent.txt")
	if err != storage.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMockStorageService_Delete(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("test.txt", []byte("test content"))

	err := mock.Delete(ctx, "test.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.DeleteCalls) != 1 {
		t.Errorf("expected 1 delete call, got %d", len(mock.DeleteCalls))
	}

	_, ok := mock.GetFile("test.txt")
	if ok {
		t.Error("expected file to be deleted")
	}
}

func TestMockStorageService_DeleteMultiple(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("file1.txt", []byte("content1"))
	mock.SetFile("file2.txt", []byte("content2"))
	mock.SetFile("file3.txt", []byte("content3"))

	err := mock.DeleteMultiple(ctx, []string{"file1.txt", "file2.txt"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(mock.DeleteCalls) != 2 {
		t.Errorf("expected 2 delete calls, got %d", len(mock.DeleteCalls))
	}
}

func TestMockStorageService_Exists(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("test.txt", []byte("test content"))

	exists, err := mock.Exists(ctx, "test.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if !exists {
		t.Error("expected file to exist")
	}

	exists, err = mock.Exists(ctx, "nonexistent.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if exists {
		t.Error("expected file to not exist")
	}
}

func TestMockStorageService_PresignedURL(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	url, err := mock.PresignedURL(ctx, "test.txt", time.Hour)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if url == "" {
		t.Error("expected URL to be returned")
	}

	if len(mock.PresignedURLCalls) != 1 {
		t.Errorf("expected 1 presigned URL call, got %d", len(mock.PresignedURLCalls))
	}
}

func TestMockStorageService_PresignedUploadURL(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	url, err := mock.PresignedUploadURL(ctx, "test.txt", time.Hour)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if url == "" {
		t.Error("expected URL to be returned")
	}
}

func TestMockStorageService_Copy(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("source.txt", []byte("source content"))

	err := mock.Copy(ctx, "source.txt", "dest.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	data, ok := mock.GetFile("dest.txt")
	if !ok {
		t.Fatal("expected destination file to exist")
	}

	if string(data) != "source content" {
		t.Errorf("expected 'source content', got '%s'", string(data))
	}
}

func TestMockStorageService_CopyNotFound(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	err := mock.Copy(ctx, "nonexistent.txt", "dest.txt")
	if err != storage.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMockStorageService_GetMetadata(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("test.txt", []byte("test content"))

	metadata, err := mock.GetMetadata(ctx, "test.txt")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if metadata["content-type"] == "" {
		t.Error("expected content-type to be set")
	}
}

func TestMockStorageService_GetMetadataNotFound(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	_, err := mock.GetMetadata(ctx, "nonexistent.txt")
	if err != storage.ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestMockStorageService_List(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("folder/file1.txt", []byte("content1"))
	mock.SetFile("folder/file2.txt", []byte("content2"))
	mock.SetFile("other/file3.txt", []byte("content3"))

	keys, err := mock.List(ctx, "folder/")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestMockStorageService_ListAll(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("file1.txt", []byte("content1"))
	mock.SetFile("file2.txt", []byte("content2"))

	keys, err := mock.List(ctx, "")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(keys) != 2 {
		t.Errorf("expected 2 keys, got %d", len(keys))
	}
}

func TestMockStorageService_Reset(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.UploadBytes(ctx, "test.txt", []byte("content"), "text/plain")
	mock.DownloadBytes(ctx, "test.txt")
	mock.Delete(ctx, "test.txt")

	mock.Reset()

	if len(mock.UploadCalls) != 0 {
		t.Error("expected upload calls to be reset")
	}

	if len(mock.DownloadCalls) != 0 {
		t.Error("expected download calls to be reset")
	}

	if len(mock.DeleteCalls) != 0 {
		t.Error("expected delete calls to be reset")
	}
}

func TestMockStorageService_SetGetFile(t *testing.T) {
	mock := NewMockStorageService()

	mock.SetFile("test.txt", []byte("test content"))

	data, ok := mock.GetFile("test.txt")
	if !ok {
		t.Fatal("expected file to exist")
	}

	if string(data) != "test content" {
		t.Errorf("expected 'test content', got '%s'", string(data))
	}
}

func TestAssertFileUploaded(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.UploadBytes(ctx, "test.txt", []byte("content"), "text/plain")

	AssertFileUploaded(t, mock, "test.txt")
}

func TestAssertFileDeleted(t *testing.T) {
	mock := NewMockStorageService()
	ctx := context.Background()

	mock.SetFile("test.txt", []byte("content"))
	mock.Delete(ctx, "test.txt")

	AssertFileDeleted(t, mock, "test.txt")
}

func TestStorageErrors(t *testing.T) {
	tests := []struct {
		name string
		err  *storage.Error
		code string
		msg  string
	}{
		{"not found", storage.ErrNotFound, "NOT_FOUND", "file not found"},
		{"upload failed", storage.ErrUploadFailed, "UPLOAD_FAILED", "upload failed"},
		{"download failed", storage.ErrDownloadFailed, "DOWNLOAD_FAILED", "download failed"},
		{"delete failed", storage.ErrDeleteFailed, "DELETE_FAILED", "delete failed"},
		{"invalid key", storage.ErrInvalidKey, "INVALID_KEY", "invalid key"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Code != tt.code {
				t.Errorf("expected code '%s', got '%s'", tt.code, tt.err.Code)
			}

			if tt.err.Message != tt.msg {
				t.Errorf("expected message '%s', got '%s'", tt.msg, tt.err.Message)
			}

			if tt.err.Error() != tt.msg {
				t.Errorf("expected Error() to return '%s', got '%s'", tt.msg, tt.err.Error())
			}
		})
	}
}

func TestConfig(t *testing.T) {
	cfg := &storage.Config{
		Provider:     storage.ProviderS3,
		Bucket:       "test-bucket",
		Region:       "us-east-1",
		Endpoint:     "https://s3.amazonaws.com",
		AccessKey:    "test-key",
		SecretKey:    "test-secret",
		UseSSL:       true,
		LocalBaseDir: "/tmp/storage",
	}

	if cfg.Provider != storage.ProviderS3 {
		t.Errorf("expected provider '%s', got '%s'", storage.ProviderS3, cfg.Provider)
	}

	if cfg.Bucket != "test-bucket" {
		t.Errorf("expected bucket 'test-bucket', got '%s'", cfg.Bucket)
	}
}

func TestProviderConstants(t *testing.T) {
	if storage.ProviderS3 != "s3" {
		t.Errorf("expected ProviderS3 to be 's3', got '%s'", storage.ProviderS3)
	}

	if storage.ProviderLocal != "local" {
		t.Errorf("expected ProviderLocal to be 'local', got '%s'", storage.ProviderLocal)
	}
}

func TestUploadOptions(t *testing.T) {
	opts := &storage.UploadOptions{}

	storage.WithContentType("text/plain")(opts)
	storage.WithContentEncoding("gzip")(opts)
	storage.WithMetadata(map[string]string{"key": "value"})(opts)
	storage.WithCacheControl("max-age=3600")(opts)

	if opts.ContentType != "text/plain" {
		t.Errorf("expected content type 'text/plain', got '%s'", opts.ContentType)
	}

	if opts.ContentEncoding != "gzip" {
		t.Errorf("expected content encoding 'gzip', got '%s'", opts.ContentEncoding)
	}

	if opts.Metadata["key"] != "value" {
		t.Error("expected metadata key to be 'value'")
	}

	if opts.CacheControl != "max-age=3600" {
		t.Errorf("expected cache control 'max-age=3600', got '%s'", opts.CacheControl)
	}
}

func TestFileInfo(t *testing.T) {
	now := time.Now()

	info := storage.FileInfo{
		Key:          "test.txt",
		Size:         1024,
		ContentType:  "text/plain",
		LastModified: now,
		ETag:         "abc123",
	}

	if info.Key != "test.txt" {
		t.Errorf("expected key 'test.txt', got '%s'", info.Key)
	}

	if info.Size != 1024 {
		t.Errorf("expected size 1024, got %d", info.Size)
	}

	if info.ContentType != "text/plain" {
		t.Errorf("expected content type 'text/plain', got '%s'", info.ContentType)
	}
}

func TestMockStorageService_CustomBehavior(t *testing.T) {
	mock := NewMockStorageService()

	mock.UploadFunc = func(ctx context.Context, key string, reader io.Reader, contentType string) error {
		return storage.ErrUploadFailed
	}

	ctx := context.Background()
	err := mock.Upload(ctx, "test.txt", bytes.NewReader([]byte("test")), "text/plain")

	if err != storage.ErrUploadFailed {
		t.Errorf("expected ErrUploadFailed, got %v", err)
	}
}

func TestMockStorageService_CustomDownloadBehavior(t *testing.T) {
	mock := NewMockStorageService()

	mock.DownloadBytesFunc = func(ctx context.Context, key string) ([]byte, error) {
		return []byte("custom content"), nil
	}

	ctx := context.Background()
	data, err := mock.DownloadBytes(ctx, "test.txt")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if string(data) != "custom content" {
		t.Errorf("expected 'custom content', got '%s'", string(data))
	}
}
