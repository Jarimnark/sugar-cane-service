package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/Jarimnark/sugar-cane-service/internal/config"
)

type Client struct {
	mc       *minio.Client
	bucket   string
	endpoint string
	secure   bool
}

func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	u, err := url.Parse(cfg.S3Endpoint)
	if err != nil {
		return nil, fmt.Errorf("parse S3_ENDPOINT: %w", err)
	}
	host := u.Host
	if host == "" {
		host = cfg.S3Endpoint
	}
	secure := u.Scheme == "https"

	mc, err := minio.New(host, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: secure,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	c := &Client{mc: mc, bucket: cfg.S3Bucket, endpoint: strings.TrimRight(cfg.S3Endpoint, "/"), secure: secure}
	if err := c.ensureBucket(ctx); err != nil {
		return nil, err
	}
	return c, nil
}

func (c *Client) ensureBucket(ctx context.Context) error {
	exists, err := c.mc.BucketExists(ctx, c.bucket)
	if err != nil {
		return fmt.Errorf("check bucket: %w", err)
	}
	if !exists {
		if err := c.mc.MakeBucket(ctx, c.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("make bucket: %w", err)
		}
	}
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, c.bucket)
	if err := c.mc.SetBucketPolicy(ctx, c.bucket, policy); err != nil {
		return fmt.Errorf("set bucket policy: %w", err)
	}
	return nil
}

func (c *Client) PutObject(ctx context.Context, prefix string, reader io.Reader, size int64, contentType string) (string, error) {
	ext := extensionFromContentType(contentType)
	key := fmt.Sprintf("%s/%s%s", strings.Trim(prefix, "/"), uuid.NewString(), ext)
	_, err := c.mc.PutObject(ctx, c.bucket, key, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}
	return fmt.Sprintf("%s/%s/%s", c.endpoint, c.bucket, key), nil
}

func extensionFromContentType(ct string) string {
	switch ct {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
