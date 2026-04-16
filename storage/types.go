package storage

import (
	"context"
	"io"
	"time"
)

type Service interface {
	Upload(ctx context.Context, key string, reader io.Reader, contentType string) error
	UploadBytes(ctx context.Context, key string, data []byte, contentType string) error
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	DownloadBytes(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	DeleteMultiple(ctx context.Context, keys []string) error
	Exists(ctx context.Context, key string) (bool, error)
	PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	PresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error)
	Copy(ctx context.Context, srcKey, dstKey string) error
	GetMetadata(ctx context.Context, key string) (map[string]string, error)
	List(ctx context.Context, prefix string) ([]string, error)
}

type FileInfo struct {
	Key          string
	Size         int64
	ContentType  string
	LastModified time.Time
	ETag         string
}

type Config struct {
	Provider     string
	Bucket       string
	Region       string
	Endpoint     string
	AccessKey    string
	SecretKey    string
	UseSSL       bool
	LocalBaseDir string
}

const (
	ProviderS3    = "s3"
	ProviderLocal = "local"
)

type UploadOptions struct {
	ContentType     string
	ContentEncoding string
	Metadata        map[string]string
	CacheControl    string
}

type UploadOption func(*UploadOptions)

func WithContentType(ct string) UploadOption {
	return func(o *UploadOptions) {
		o.ContentType = ct
	}
}

func WithContentEncoding(ce string) UploadOption {
	return func(o *UploadOptions) {
		o.ContentEncoding = ce
	}
}

func WithMetadata(m map[string]string) UploadOption {
	return func(o *UploadOptions) {
		o.Metadata = m
	}
}

func WithCacheControl(cc string) UploadOption {
	return func(o *UploadOptions) {
		o.CacheControl = cc
	}
}

var (
	ErrNotFound       = &Error{Code: "NOT_FOUND", Message: "file not found"}
	ErrUploadFailed   = &Error{Code: "UPLOAD_FAILED", Message: "upload failed"}
	ErrDownloadFailed = &Error{Code: "DOWNLOAD_FAILED", Message: "download failed"}
	ErrDeleteFailed   = &Error{Code: "DELETE_FAILED", Message: "delete failed"}
	ErrInvalidKey     = &Error{Code: "INVALID_KEY", Message: "invalid key"}
)

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return e.Message
}
