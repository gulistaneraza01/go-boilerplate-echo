package aws

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/gulistaneraza01/go-boilerplate-echo/internal/config"
)

type Storage struct {
	Client  *s3.Client
	presign *s3.PresignClient
	bucket  string
}

// NewS3Client builds an S3 client from the shared AWS config
func NewS3Client(awsCfg aws.Config, cfg config.S3Config) *s3.Client {
	return s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		if cfg.EndpointURL != "" {
			o.BaseEndpoint = aws.String(cfg.EndpointURL)
			o.UsePathStyle = true // required by most S3-compatible stores
		}
	})
}

// NewStorage wraps client for bucket.
func NewStorage(client *s3.Client, bucket string) *Storage {
	return &Storage{Client: client, presign: s3.NewPresignClient(client), bucket: bucket}
}

// Upload streams file to the bucket under key (a random UUID if empty) and returns the key.
func (s *Storage) Upload(ctx context.Context, key string, file io.ReadSeeker) (string, error) {
	head := make([]byte, 512)
	n, err := io.ReadFull(file, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return "", fmt.Errorf("failed to read file: %w", err)
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", fmt.Errorf("failed to rewind file: %w", err)
	}

	key = keyOrUUID(key)
	_, err = s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(http.DetectContentType(head[:n])),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}
	return key, nil
}

// CreatePresignedUploadURL returns a temporary PUT link for a direct client upload under key (a random UUID if empty).
func (s *Storage) CreatePresignedUploadURL(ctx context.Context, key, contentType string, size int64, ttl time.Duration) (string, string, error) {
	if contentType == "" {
		return "", "", errors.New("content type is required")
	}
	if size <= 0 {
		return "", "", fmt.Errorf("invalid upload size %d", size)
	}

	key = keyOrUUID(key)
	req, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:        aws.String(s.bucket),
		Key:           aws.String(key),
		ContentType:   aws.String(contentType),
		ContentLength: aws.Int64(size),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", "", fmt.Errorf("failed to presign upload: %w", err)
	}
	return req.URL, key, nil
}

// CreatePresignedURL returns a temporary GET link to key, valid for ttl.
func (s *Storage) CreatePresignedURL(ctx context.Context, key string, ttl time.Duration) (string, error) {
	req, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(ttl))
	if err != nil {
		return "", fmt.Errorf("failed to presign %s: %w", key, err)
	}
	return req.URL, nil
}

// Delete removes key from the bucket. Deleting a missing key is not an error.
func (s *Storage) Delete(ctx context.Context, key string) error {
	if _, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}); err != nil {
		return fmt.Errorf("failed to delete object %s: %w", key, err)
	}
	return nil
}

// keyOrUUID returns key, or a random UUID when key is empty.
func keyOrUUID(key string) string {
	if key == "" {
		return uuid.NewString()
	}
	return key
}
