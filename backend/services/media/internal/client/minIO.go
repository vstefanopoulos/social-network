package client

import (
	"context"
	"fmt"
	"net/url"
	ce "social-network/shared/go/commonerrors"
	"time"

	"github.com/minio/minio-go/v7"
)

func (c *Clients) GenerateDownloadURL(
	ctx context.Context,
	bucket string,
	objectKey string,
	expiry time.Duration,
) (*url.URL, error) {
	errMsg := fmt.Sprintf("S3 client: generate download: file bucket: %v object key: %v", bucket, objectKey)

	client := c.MinIOClient

	// Only for development
	if c.PublicMinIOClient != nil {
		client = c.PublicMinIOClient
	}

	url, err := client.PresignedGetObject(
		ctx,
		bucket,
		objectKey,
		expiry,
		nil,
	)

	if err != nil {
		return nil, ce.Wrap(ce.ErrInternal, err, errMsg)
	}

	return url, nil
}

func (c *Clients) GenerateUploadURL(
	ctx context.Context,
	bucket string,
	objectKey string,
	expiry time.Duration,
) (*url.URL, error) {
	errMsg := fmt.Sprintf("S3 client: generate upload: file bucket: %v object key: %v", bucket, objectKey)

	client := c.MinIOClient

	// Only for development
	if c.PublicMinIOClient != nil {
		client = c.PublicMinIOClient
	}

	url, err := client.PresignedPutObject(
		ctx,
		bucket,
		objectKey,
		expiry,
	)

	if err != nil {
		return nil, ce.Wrap(ce.ErrInternal, err, errMsg)
	}

	return url, nil
}

func (c *Clients) DeleteFile(ctx context.Context,
	bucket string,
	objectKey string,
) error {
	return c.MinIOClient.RemoveObject(
		ctx,
		bucket,
		objectKey,
		minio.RemoveObjectOptions{},
	)
}
