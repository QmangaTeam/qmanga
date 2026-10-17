package s3

import (
	"context"
	"errors"
	"fmt"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	minioClient  *minio.Client
	mediaBucket  string
	sourceBucket string
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
