package client

import (
	"bytes"
	"context"
	"io"
	"social-network/services/media/internal/configs"
	ct "social-network/shared/go/ct"

	"github.com/minio/minio-go/v7"
)

type Clients struct {
	Configs           configs.FileService
	MinIOClient       *minio.Client // replace with proper S3 service on prod
	PublicMinIOClient *minio.Client // only for development
	Validator         Validator
	ImageConvertor    ImageConvertor
}

type Validator interface {
	// ValidateImage checks that the provided image meets size, format, and dimension constraints.
	// Returns an error if the image is invalid or unsupported.
	ValidateImage(ctx context.Context, r io.Reader) error
}

type ImageConvertor interface {
	// ConvertImageToVariant resizes an image to the given variant and encodes it as WebP.
	ConvertImageToVariant(
		buf []byte, variant ct.FileVariant,
	) (out bytes.Buffer, err error)
}
