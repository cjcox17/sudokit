package storage

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

type LocalService struct {
	baseDir string
}

func NewLocalService(baseDir string) (*LocalService, error) {
	if baseDir == "" {
		baseDir = "storage"
	}

	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	slog.Info("Local storage service initialized", "baseDir", baseDir)
	return &LocalService{baseDir: baseDir}, nil
}

func (s *LocalService) fullPath(key string) string {
	return filepath.Join(s.baseDir, key)
}

func (s *LocalService) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	if key == "" {
		return ErrInvalidKey
	}

	fullPath := s.fullPath(key)
	dir := filepath.Dir(fullPath)

	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}
	defer file.Close()

	_, err = io.Copy(file, reader)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	slog.Debug("File uploaded", "key", key)
	return nil
}

func (s *LocalService) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	return s.Upload(ctx, key, bytes.NewReader(data), contentType)
}

func (s *LocalService) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	fullPath := s.fullPath(key)

	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	return file, nil
}

func (s *LocalService) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	reader, err := s.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (s *LocalService) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	fullPath := s.fullPath(key)

	err := os.Remove(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	slog.Debug("File deleted", "key", key)
	return nil
}

func (s *LocalService) DeleteMultiple(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := s.Delete(ctx, key); err != nil {
			return err
		}
	}
	return nil
}

func (s *LocalService) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	fullPath := s.fullPath(key)

	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *LocalService) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}

	return fmt.Sprintf("/storage/%s", key), nil
}

func (s *LocalService) PresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}

	return fmt.Sprintf("/storage/upload/%s", key), nil
}

func (s *LocalService) Copy(ctx context.Context, srcKey, dstKey string) error {
	if srcKey == "" || dstKey == "" {
		return ErrInvalidKey
	}

	srcPath := s.fullPath(srcKey)
	dstPath := s.fullPath(dstKey)

	dir := filepath.Dir(dstPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrNotFound
		}
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	slog.Debug("File copied", "src", srcKey, "dst", dstKey)
	return nil
}

func (s *LocalService) GetMetadata(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	fullPath := s.fullPath(key)

	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	metadata := make(map[string]string)
	metadata["size"] = fmt.Sprintf("%d", info.Size())
	metadata["last-modified"] = info.ModTime().Format(time.RFC3339)
	metadata["content-type"] = detectContentType(key, nil)

	return metadata, nil
}

func (s *LocalService) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	base := s.baseDir
	if prefix != "" {
		base = filepath.Join(s.baseDir, prefix)
	}

	err := filepath.Walk(base, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}

		relPath, err := filepath.Rel(s.baseDir, path)
		if err != nil {
			return err
		}

		keys = append(keys, relPath)
		return nil
	})

	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	return keys, nil
}
