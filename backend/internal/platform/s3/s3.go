package s3

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	minioClient  *minio.Client
	mediaBucket  string
	sourceBucket string
}

func (clientInstance *Client) put(ctx context.Context, bucket, key string, reader io.Reader, size int64, mime string) error {
	_, failure := clientInstance.minioClient.PutObject(ctx, bucket, key, reader, size, minio.PutObjectOptions{ContentType: mime, NumThreads: 1, PartSize: 8 << 20, DisableMultipart: size <= 32<<20})
	return failure
}
func (clientInstance *Client) PutSource(ctx context.Context, key string, reader io.Reader, size int64, mime string) error {
	return clientInstance.put(ctx, clientInstance.sourceBucket, key, reader, size, mime)
}
func (clientInstance *Client) PutDerivative(ctx context.Context, key string, reader io.Reader, size int64, mime string) error {
	return clientInstance.put(ctx, clientInstance.mediaBucket, key, reader, size, mime)
}
func (clientInstance *Client) OpenSource(ctx context.Context, key string) (io.ReadCloser, error) {
	return clientInstance.open(ctx, clientInstance.sourceBucket, key, 0, -1)
}
func (clientInstance *Client) OpenDerivative(ctx context.Context, key string, offset, length int64) (io.ReadCloser, error) {
	return clientInstance.open(ctx, clientInstance.mediaBucket, key, offset, length)
}
func (clientInstance *Client) open(ctx context.Context, bucket, key string, offset, length int64) (io.ReadCloser, error) {
	options := minio.GetObjectOptions{}
	if length >= 0 {
		if failure := options.SetRange(offset, offset+length-1); failure != nil {
			return nil, failure
		}
	}
	object, failure := clientInstance.minioClient.GetObject(ctx, bucket, key, options)
	if failure != nil {
		return nil, failure
	}
	if _, failure = object.Stat(); failure != nil {
		object.Close()
		return nil, failure
	}
	return object, nil
}

func New(endpoint string, accessKey string, secretKey string, mediaBucket string, sourceBucket string, useSSL bool) (*Client, error) {
	if endpoint == "" {
		return nil, errors.New("s3 endpoint must not be empty")
	}

	minioOptions := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	}

	minioClient, initError := minio.New(endpoint, minioOptions)
	if initError != nil {
		return nil, fmt.Errorf("initializing minio client: %w", initError)
	}

	clientInstance := &Client{
		minioClient:  minioClient,
		mediaBucket:  mediaBucket,
		sourceBucket: sourceBucket,
	}

	return clientInstance, nil
}

func (clientInstance *Client) Ping(contextInstance context.Context) error {
	if clientInstance == nil || clientInstance.minioClient == nil {
		return errors.New("s3 client is not initialized")
	}

	targetBuckets := []string{clientInstance.mediaBucket, clientInstance.sourceBucket}
	for _, bucketName := range targetBuckets {
		if bucketName == "" {
			return errors.New("configured bucket name is empty")
		}
		bucketExists, checkError := clientInstance.minioClient.BucketExists(contextInstance, bucketName)
		if checkError != nil {
			return fmt.Errorf("checking bucket accessibility %s: %w", bucketName, checkError)
		}
		if !bucketExists {
			return fmt.Errorf("configured bucket does not exist: %s", bucketName)
		}
	}

	return nil
}

func (clientInstance *Client) EnsureBuckets(contextInstance context.Context) error {
	if clientInstance == nil || clientInstance.minioClient == nil {
		return errors.New("s3 client is not initialized")
	}

	targetBuckets := []string{clientInstance.mediaBucket, clientInstance.sourceBucket}
	for _, bucketName := range targetBuckets {
		if bucketName == "" {
			return errors.New("cannot ensure empty bucket name")
		}
		bucketExists, checkError := clientInstance.minioClient.BucketExists(contextInstance, bucketName)
		if checkError != nil {
			return fmt.Errorf("checking bucket %s: %w", bucketName, checkError)
		}
		if !bucketExists {
			makeError := clientInstance.minioClient.MakeBucket(contextInstance, bucketName, minio.MakeBucketOptions{})
			if makeError != nil {
				return fmt.Errorf("creating bucket %s: %w", bucketName, makeError)
			}
		}
	}

	return nil
}

func (clientInstance *Client) DeleteDerivative(ctx context.Context, key string) error {
	return clientInstance.minioClient.RemoveObject(ctx, clientInstance.mediaBucket, key, minio.RemoveObjectOptions{})
}
