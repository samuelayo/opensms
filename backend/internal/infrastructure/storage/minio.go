package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/samuelayo/opensms/internal/infrastructure/config"
)

// Storage defines the interface for object storage operations
type Storage interface {
	Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error
	Download(ctx context.Context, bucket, objectName string) (io.ReadCloser, error)
	Delete(ctx context.Context, bucket, objectName string) error
	GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error)
	CreateBucket(ctx context.Context, bucket string) error
	BucketExists(ctx context.Context, bucket string) (bool, error)
}

// MinIOClient wraps minio.Client with additional functionality
type MinIOClient struct {
	client *minio.Client
	config config.MinIOConfig
}

// NewMinIOClient creates a new MinIO/S3 client
func NewMinIOClient(cfg config.MinIOConfig) (*MinIOClient, error) {
	// Initialize MinIO client
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	mc := &MinIOClient{
		client: client,
		config: cfg,
	}

	// Ensure default bucket exists
	ctx := context.Background()
	exists, err := mc.BucketExists(ctx, cfg.Bucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		if err := mc.CreateBucket(ctx, cfg.Bucket); err != nil {
			return nil, fmt.Errorf("failed to create default bucket: %w", err)
		}
	}

	return mc, nil
}

// Upload uploads an object to storage
func (m *MinIOClient) Upload(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) error {
	opts := minio.PutObjectOptions{
		ContentType: contentType,
	}

	_, err := m.client.PutObject(ctx, bucket, objectName, reader, size, opts)
	if err != nil {
		return fmt.Errorf("failed to upload object: %w", err)
	}

	return nil
}

// Download downloads an object from storage
func (m *MinIOClient) Download(ctx context.Context, bucket, objectName string) (io.ReadCloser, error) {
	object, err := m.client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download object: %w", err)
	}

	return object, nil
}

// Delete deletes an object from storage
func (m *MinIOClient) Delete(ctx context.Context, bucket, objectName string) error {
	opts := minio.RemoveObjectOptions{}
	if err := m.client.RemoveObject(ctx, bucket, objectName, opts); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}

	return nil
}

// GetPresignedURL generates a presigned URL for temporary access
func (m *MinIOClient) GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return url.String(), nil
}

// CreateBucket creates a new bucket
func (m *MinIOClient) CreateBucket(ctx context.Context, bucket string) error {
	opts := minio.MakeBucketOptions{
		Region: m.config.Region,
	}

	if err := m.client.MakeBucket(ctx, bucket, opts); err != nil {
		return fmt.Errorf("failed to create bucket: %w", err)
	}

	return nil
}

// BucketExists checks if a bucket exists
func (m *MinIOClient) BucketExists(ctx context.Context, bucket string) (bool, error) {
	exists, err := m.client.BucketExists(ctx, bucket)
	if err != nil {
		return false, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	return exists, nil
}

// ListObjects lists objects in a bucket with prefix
func (m *MinIOClient) ListObjects(ctx context.Context, bucket, prefix string) ([]string, error) {
	opts := minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	}

	var objects []string
	for object := range m.client.ListObjects(ctx, bucket, opts) {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list objects: %w", object.Err)
		}
		objects = append(objects, object.Key)
	}

	return objects, nil
}

// Helper functions for common storage operations

// UploadProfilePicture uploads a user profile picture
func (m *MinIOClient) UploadProfilePicture(ctx context.Context, userID string, reader io.Reader, size int64, contentType string) (string, error) {
	objectName := fmt.Sprintf("profiles/%s/avatar", userID)
	if err := m.Upload(ctx, m.config.Bucket, objectName, reader, size, contentType); err != nil {
		return "", err
	}
	return objectName, nil
}

// UploadDocument uploads a document (assignment, report, etc.)
func (m *MinIOClient) UploadDocument(ctx context.Context, category, id string, filename string, reader io.Reader, size int64, contentType string) (string, error) {
	objectName := fmt.Sprintf("documents/%s/%s/%s", category, id, filename)
	if err := m.Upload(ctx, m.config.Bucket, objectName, reader, size, contentType); err != nil {
		return "", err
	}
	return objectName, nil
}
