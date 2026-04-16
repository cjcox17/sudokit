package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type S3Service struct {
	client *s3.Client
	bucket string
	region string
}

func NewS3Service(cfg *Config) (*S3Service, error) {
	if cfg.Bucket == "" {
		return nil, fmt.Errorf("bucket name is required")
	}

	if cfg.Region == "" {
		cfg.Region = "us-east-1"
	}

	var opts []func(*config.LoadOptions) error

	if cfg.AccessKey != "" && cfg.SecretKey != "" {
		opts = append(opts, config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, ""),
		))
	}

	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	awsCfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	clientOpts := []func(*s3.Options){}
	if cfg.Endpoint != "" {
		clientOpts = append(clientOpts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			if strings.Contains(cfg.Endpoint, "localhost") || strings.Contains(cfg.Endpoint, "127.0.0.1") {
				o.UsePathStyle = true
			}
		})
	}

	client := s3.NewFromConfig(awsCfg, clientOpts...)

	slog.Info("S3 storage service initialized", "bucket", cfg.Bucket, "region", cfg.Region)
	return &S3Service{
		client: client,
		bucket: cfg.Bucket,
		region: cfg.Region,
	}, nil
}

func (s *S3Service) Upload(ctx context.Context, key string, reader io.Reader, contentType string) error {
	if key == "" {
		return ErrInvalidKey
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        reader,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		slog.Error("Failed to upload file", "key", key, "error", err)
		return fmt.Errorf("%w: %v", ErrUploadFailed, err)
	}

	slog.Debug("File uploaded", "key", key, "bucket", s.bucket)
	return nil
}

func (s *S3Service) UploadBytes(ctx context.Context, key string, data []byte, contentType string) error {
	return s.Upload(ctx, key, bytes.NewReader(data), contentType)
}

func (s *S3Service) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	resp, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, ErrNotFound
		}
		slog.Error("Failed to download file", "key", key, "error", err)
		return nil, fmt.Errorf("%w: %v", ErrDownloadFailed, err)
	}

	return resp.Body, nil
}

func (s *S3Service) DownloadBytes(ctx context.Context, key string) ([]byte, error) {
	reader, err := s.Download(ctx, key)
	if err != nil {
		return nil, err
	}
	defer reader.Close()

	return io.ReadAll(reader)
}

func (s *S3Service) Delete(ctx context.Context, key string) error {
	if key == "" {
		return ErrInvalidKey
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		slog.Error("Failed to delete file", "key", key, "error", err)
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	slog.Debug("File deleted", "key", key)
	return nil
}

func (s *S3Service) DeleteMultiple(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(key)}
	}

	_, err := s.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(s.bucket),
		Delete: &types.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteFailed, err)
	}

	slog.Debug("Files deleted", "count", len(keys))
	return nil
}

func (s *S3Service) Exists(ctx context.Context, key string) (bool, error) {
	if key == "" {
		return false, ErrInvalidKey
	}

	_, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (s *S3Service) PresignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}

	presignClient := s3.NewPresignClient(s.client, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	req, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return req.URL, nil
}

func (s *S3Service) PresignedUploadURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	if key == "" {
		return "", ErrInvalidKey
	}

	presignClient := s3.NewPresignClient(s.client, func(opts *s3.PresignOptions) {
		opts.Expires = expiry
	})

	req, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned upload URL: %w", err)
	}

	return req.URL, nil
}

func (s *S3Service) Copy(ctx context.Context, srcKey, dstKey string) error {
	if srcKey == "" || dstKey == "" {
		return ErrInvalidKey
	}

	copySource := fmt.Sprintf("%s/%s", s.bucket, srcKey)

	_, err := s.client.CopyObject(ctx, &s3.CopyObjectInput{
		Bucket:     aws.String(s.bucket),
		Key:        aws.String(dstKey),
		CopySource: aws.String(copySource),
	})
	if err != nil {
		return fmt.Errorf("failed to copy object: %w", err)
	}

	slog.Debug("File copied", "src", srcKey, "dst", dstKey)
	return nil
}

func (s *S3Service) GetMetadata(ctx context.Context, key string) (map[string]string, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}

	resp, err := s.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		var noSuchKey *types.NoSuchKey
		if errors.As(err, &noSuchKey) {
			return nil, ErrNotFound
		}
		return nil, err
	}

	metadata := make(map[string]string)
	if resp.ContentType != nil {
		metadata["content-type"] = *resp.ContentType
	}
	if resp.ContentLength != nil {
		metadata["size"] = fmt.Sprintf("%d", *resp.ContentLength)
	}
	if resp.LastModified != nil {
		metadata["last-modified"] = resp.LastModified.Format(time.RFC3339)
	}
	if resp.ETag != nil {
		metadata["etag"] = *resp.ETag
	}

	for k, v := range resp.Metadata {
		metadata[k] = v
	}

	return metadata, nil
}

func (s *S3Service) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string

	paginator := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(s.bucket),
		Prefix: aws.String(prefix),
	})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, obj := range page.Contents {
			if obj.Key != nil {
				keys = append(keys, *obj.Key)
			}
		}
	}

	return keys, nil
}
