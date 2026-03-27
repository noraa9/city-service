package minio

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path"
	"strings"

	"city-service/internal/config"

	miniolib "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Client wraps the official MinIO client and adds a couple project-specific helpers.
//
// Why a wrapper:
// - services should not know MinIO policy JSON, URL formatting, etc.
// - easier to mock in tests by creating a small interface later
type Client struct {
	s3     *miniolib.Client
	scheme string // "http" or "https" (used for building public URLs)
}

// NewMinioClient connects to MinIO and ensures the bucket exists and is public.
func NewMinioClient(cfg config.Config) (*Client, error) {
	if cfg.MinioEndpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT is required")
	}
	if cfg.MinioAccessKey == "" || cfg.MinioSecretKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY and MINIO_SECRET_KEY are required")
	}
	if cfg.MinioBucket == "" {
		return nil, fmt.Errorf("MINIO_BUCKET is required")
	}

	cli, err := miniolib.New(cfg.MinioEndpoint, &miniolib.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	c := &Client{
		s3:     cli,
		scheme: "http",
	}
	if cfg.MinioUseSSL {
		c.scheme = "https"
	}

	// Ensure bucket exists and is public readable so the frontend can load images by URL.
	ctx := context.Background()
	if err := c.ensureBucketPublic(ctx, cfg.MinioBucket); err != nil {
		return nil, err
	}

	return c, nil
}

// UploadFile uploads a file to the given bucket and returns a public URL.
//
// fileName should already be unique (we typically use uuid + original extension).
func (c *Client) UploadFile(ctx context.Context, bucketName, fileName string, file io.Reader, size int64, contentType string) (string, error) {
	_, err := c.s3.PutObject(ctx, bucketName, fileName, file, size, miniolib.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload to minio: %w", err)
	}
	return c.GetFileURL(bucketName, fileName), nil
}

// GetFileURL builds a direct URL to the object (bucket is public-read).
//
// Example: http://localhost:9000/city-service/filename.jpg
func (c *Client) GetFileURL(bucketName, fileName string) string {
	// MinIO serves objects at: {scheme}://{endpoint}/{bucket}/{object}
	return fmt.Sprintf("%s://%s/%s/%s", c.scheme, c.s3.EndpointURL().Host, bucketName, strings.TrimLeft(fileName, "/"))
}

func (c *Client) ensureBucketPublic(ctx context.Context, bucket string) error {
	exists, err := c.s3.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("check bucket exists: %w", err)
	}
	if !exists {
		if err := c.s3.MakeBucket(ctx, bucket, miniolib.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create bucket %q: %w", bucket, err)
		}
	}

	// Public read bucket policy:
	// - allow everyone ("Principal":"*") to perform s3:GetObject on bucket/*
	policy := map[string]any{
		"Version": "2012-10-17",
		"Statement": []map[string]any{
			{
				"Sid":       "PublicReadGetObject",
				"Effect":    "Allow",
				"Principal": "*",
				"Action":    []string{"s3:GetObject"},
				"Resource":  []string{fmt.Sprintf("arn:aws:s3:::%s/*", bucket)},
			},
		},
	}
	b, _ := json.Marshal(policy)
	if err := c.s3.SetBucketPolicy(ctx, bucket, string(b)); err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}
	return nil
}

// BuildObjectName is a convenience helper for creating unique file names.
//
// Example:
// - original: "pothole.jpg"
// - uuid:     "550e8400-e29b-41d4-a716-446655440000"
// - result:   "550e8400-e29b-41d4-a716-446655440000.jpg"
func BuildObjectName(id string, originalFilename string) string {
	ext := path.Ext(originalFilename) // includes dot, e.g. ".jpg"
	return id + ext
}
