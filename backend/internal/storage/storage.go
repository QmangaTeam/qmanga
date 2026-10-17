package storage

import (
	"context"
	"fmt"

	"qmanga/internal/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	client       *minio.Client
	mediaBucket  string
	sourceBucket string
}

func New(cfg config.Config) (*Client, error) {
	if cfg.S3Endpoint == "" {
		return &Client{}, nil
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.S3AccessKey, cfg.S3SecretKey, ""),
		Secure: cfg.S3UseSSL,
	}

	client, err := minio.New(cfg.S3Endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("initializing minio client: %w", err)
	}

	c := &Client{
		client:       client,
		mediaBucket:  cfg.S3BucketMedia,
		sourceBucket: cfg.S3BucketSource,
	}

	return c, nil
}

func (c *Client) Ping(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}
	_, err := c.client.ListBuckets(ctx)
	if err != nil {
		return fmt.Errorf("listing s3 buckets: %w", err)
	}
	return nil
}

func (c *Client) EnsureBuckets(ctx context.Context) error {
	if c == nil || c.client == nil {
		return nil
	}

	buckets := []string{c.mediaBucket, c.sourceBucket}
	for _, bucket := range buckets {
		if bucket == "" {
			continue
		}
		exists, err := c.client.BucketExists(ctx, bucket)
		if err != nil {
			return fmt.Errorf("checking bucket %s: %w", bucket, err)
		}
		if !exists {
			err = c.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
			if err != nil {
				return fmt.Errorf("creating bucket %s: %w", bucket, err)
			}
		}
	}

	return nil
}
