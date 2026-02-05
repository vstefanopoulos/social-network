package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path/filepath"
	md "social-network/services/media/internal/models"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/tags"
)

func (c *Clients) CheckValidationStatus(ctx context.Context,
	fm md.FileMeta) (bool, error) {
	errMsg := "s3 client: check validation status"
	tagging, err := c.MinIOClient.GetObjectTagging(
		ctx,
		fm.Bucket,
		fm.ObjectKey,
		minio.GetObjectTaggingOptions{},
	)
	if err != nil {
		if minio.ToErrorResponse(err).Code != "NoSuchTagSet" {
			return false, ce.Wrap(ce.ErrInternal, err, errMsg)
		}
	}

	existingTags := tagging.ToMap()
	if existingTags["validated"] == "true" {
		return true, nil
	}
	return false, nil
}

// DEPRACATED
func (c *Clients) ValidateUpload(
	ctx context.Context,
	fm md.FileMeta,
) error {
	input := fmt.Sprintf("file meta: %#v", fm)

	validated, _ := c.CheckValidationStatus(ctx, fm)
	if validated {
		return nil
	}

	fileCnstr := c.Configs.FileConstraints

	info, err := c.MinIOClient.StatObject(
		ctx,
		fm.Bucket,
		fm.ObjectKey,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return ce.Wrap(ce.ErrNotFound, err, input) // upload never completed
	}

	if err := checkSize(fm.SizeBytes, info.Size, fileCnstr.MaxImageUpload); err != nil {
		return ce.Wrap(nil, err, input)
	}

	if err := checkExt(fileCnstr.AllowedExt, fm.Filename); err != nil {
		return ce.Wrap(nil, err, input)
	}

	if err := checkMime(fileCnstr.AllowedMIMEs, fm.MimeType); err != nil {
		return ce.Wrap(nil, err, input)
	}

	obj, err := c.MinIOClient.GetObject(ctx, fm.Bucket, fm.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err)
	}
	defer obj.Close()
	if err := c.Validator.ValidateImage(ctx, obj); err != nil {
		return ce.Wrap(nil, err, input) // Validate returns customerrors type with public message
	}

	tagSet, err := tags.NewTags(map[string]string{
		"validated": "true",
	}, true,
	)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, input)
	}

	err = c.MinIOClient.PutObjectTagging(
		ctx,
		fm.Bucket,
		fm.ObjectKey,
		tagSet,
		minio.PutObjectTaggingOptions{},
	)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, input+": putObjectTagging")
	}
	return nil
}

type VariantToGenerate struct {
	Id      ct.Id
	Bucket  string
	ObjKey  string
	Variant ct.FileVariant
	Size    int64
}

// Validates file and creates all linked variants. If any part of the process returns error it
func (c *Clients) ValidateAndCreateVariants(
	ctx context.Context,
	fm md.FileMeta,
	variants []VariantToGenerate,
) *ce.Error {
	input := fmt.Sprintf("file meta: %#v", fm)

	validated, _ := c.CheckValidationStatus(ctx, fm)
	if validated {
		return nil
	}

	fileCnstr := c.Configs.FileConstraints

	info, err := c.MinIOClient.StatObject(
		ctx,
		fm.Bucket,
		fm.ObjectKey,
		minio.StatObjectOptions{},
	)
	if err != nil {
		return ce.Wrap(ce.ErrNotFound, err, input) // upload never completed
	}

	if err := checkSize(fm.SizeBytes, info.Size, fileCnstr.MaxImageUpload); err != nil {
		return ce.Wrap(nil, err, input)
	}

	if err := checkExt(fileCnstr.AllowedExt, fm.Filename); err != nil {
		return ce.Wrap(nil, err, input)
	}

	if err := checkMime(fileCnstr.AllowedMIMEs, fm.MimeType); err != nil {
		return ce.Wrap(nil, err, input)
	}

	obj, err := c.MinIOClient.GetObject(ctx, fm.Bucket, fm.ObjectKey, minio.GetObjectOptions{})
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err)
	}
	defer obj.Close()

	data, err := io.ReadAll(obj)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, "failed to read original object")
	}

	if err := c.Validator.ValidateImage(ctx, bytes.NewReader(data)); err != nil {
		return ce.Wrap(nil, err, input) // Validate returns customerrors type with public message
	}

	tele.Debug(ctx, "image validation success", "file meta", fm)

	tagSet, err := tags.NewTags(map[string]string{
		"validated": "true",
	}, true,
	)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, input)
	}

	err = c.MinIOClient.PutObjectTagging(
		ctx,
		fm.Bucket,
		fm.ObjectKey,
		tagSet,
		minio.PutObjectTaggingOptions{},
	)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, input+": putObjectTagging")
	}

	if err := c.GenVariants(ctx, data, variants); err != nil {
		return ce.Wrap(nil, err, input)
	}
	return nil
}

// Compares s3 object size with promised and max allowed size,
func checkSize(actual, promised, max int64) *ce.Error {
	if actual != promised {
		return ce.New(
			ce.ErrPermissionDenied,
			fmt.Errorf("upload size mismatch: expected=%d actual=%d", promised, actual),
		).WithPublic("promised and actual size mismatch")
	}
	if actual > max {
		return ce.Wrap(ce.ErrPermissionDenied,
			fmt.Errorf("image size %v exceedes allowed size %v", actual, max),
		).WithPublic("file too big")
	}
	return nil
}

// Compares file extention with allowed extentions.
func checkExt(allowedExt map[string]bool, filename string) *ce.Error {
	ext := strings.ToLower(filepath.Ext(filename))
	if ok := allowedExt[ext]; !ok {
		return ce.Wrap(
			ce.ErrPermissionDenied,
			fmt.Errorf("invalid file ext %v", ext),
		).WithPublic("invalid file extension")
	}
	return nil
}

// Compares mime type with allowed mime types.
func checkMime(allowedMIMEs map[string]bool, mime string) *ce.Error {
	if ok := allowedMIMEs[mime]; !ok {
		return ce.New(ce.ErrPermissionDenied,
			fmt.Errorf("unsuported mime type %v", mime),
		).WithPublic(fmt.Sprintf("unsuported mime type %v", mime))
	}
	return nil
}
