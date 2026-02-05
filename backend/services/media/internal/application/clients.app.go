package application

import (
	"context"
	"net/url"
	"social-network/services/media/internal/client"
	md "social-network/services/media/internal/models"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"time"
)

type S3Service interface {
	GenerateDownloadURL(
		ctx context.Context,
		bucket string,
		objectKey string,
		expiry time.Duration,
	) (*url.URL, error)

	GenerateUploadURL(
		ctx context.Context,
		bucket string,
		objectKey string,
		expiry time.Duration,
	) (*url.URL, error)

	ValidateUpload(
		ctx context.Context,
		upload md.FileMeta,
	) error

	ValidateAndCreateVariants(
		ctx context.Context,
		fm md.FileMeta,
		variants []client.VariantToGenerate,
	) *ce.Error

	GenerateVariant(
		ctx context.Context,
		srcBucket string,
		srcObjectKey string,
		trgBucket string,
		trgObjectKey string,
		variant ct.FileVariant,
	) (size int64, err error)

	DeleteFile(ctx context.Context,
		bucket string,
		objectKey string,
	) error
}
