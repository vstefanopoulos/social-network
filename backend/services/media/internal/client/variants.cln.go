package client

import (
	"context"
	"fmt"
	"io"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"

	"github.com/minio/minio-go/v7"
)

func (c *Clients) GenerateVariant(
	ctx context.Context,
	srcBucket string,
	srcObjectKey string,
	trgBucket string,
	trgObjectKey string,
	variant ct.FileVariant,
) (size int64, err error) {

	obj, err := c.MinIOClient.GetObject(ctx,
		srcBucket, srcObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return 0, ce.Wrap(ce.ErrNotFound, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return 0, ce.Wrap(ce.ErrInternal, err, "failed to read original object")
	}

	outBuf, err := c.ImageConvertor.ConvertImageToVariant(data, variant)

	info, err := c.MinIOClient.PutObject(
		ctx,
		trgBucket,
		trgObjectKey,
		&outBuf,
		int64(outBuf.Len()),
		minio.PutObjectOptions{
			ContentType: "image/webp",
		},
	)
	size = info.Size
	if err != nil {
		return 0, ce.Wrap(ce.ErrInternal, err)
	}
	return size, nil
}

// Generates all variants in 'variants' argument puts them to file service
// and updates the Size field in VariantToGenerate.
func (c *Clients) GenVariants(ctx context.Context, data []byte, variants []VariantToGenerate) *ce.Error {
	for i := range variants {
		outBuf, err := c.ImageConvertor.ConvertImageToVariant(data, variants[i].Variant)
		if err != nil {
			return ce.New(ce.ErrInternal, err, fmt.Sprintf("failed to generate variant: %#v", variants[i]))
		}
		info, err := c.MinIOClient.PutObject(
			ctx,
			variants[i].Bucket,
			variants[i].ObjKey,
			&outBuf,
			int64(outBuf.Len()),
			minio.PutObjectOptions{
				ContentType: "image/webp",
			},
		)
		variants[i].Size = info.Size
		if err != nil {
			return ce.New(ce.ErrInternal, err, fmt.Sprintf("failed to generate variant: %#v", variants[i]))
		}
		tele.Debug(ctx, "GenVariants: variant created", "upload info", info, "variant", variants[i])
	}
	return nil
}
