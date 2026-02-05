package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/url"
	"social-network/services/media/internal/db/dbservice"
	"social-network/services/media/internal/mapping"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"
	"time"

	"github.com/google/uuid"
)

type UploadImageReq struct {
	Filename   string
	MimeType   string
	SizeBytes  int64
	Visibility ct.FileVisibility
}

// Provides a fileId and an upload url targeted on bucket Originals defined on configs.
// Creates all variant entries provided in []variants for workers to later
// create asynchronously the compressed files.
func (m *MediaService) UploadImage(ctx context.Context,
	req UploadImageReq,
	exp time.Duration,
	variants []ct.FileVariant,
) (fileId ct.Id, upUrl string, err error) {
	input := fmt.Sprintf("req: %#v, variants: %v", req, variants)

	if err := m.validateUploadRequest(
		req,
		exp,
		variants,
	); err != nil {
		return 0, "", ce.Wrap(ce.ErrInvalidArgument, err)
	}

	objectKey := uuid.NewString()
	orignalsBucket := m.Cfgs.FileService.Buckets.Originals
	variantsBucket := m.Cfgs.FileService.Buckets.Variants
	var url *url.URL

	errTx := m.txRunner.RunTx(ctx,
		func(tx *dbservice.Queries) error {
			fileId, err = tx.CreateFile(ctx, dbservice.File{
				Filename:   req.Filename,
				MimeType:   req.MimeType,
				SizeBytes:  req.SizeBytes,
				Visibility: req.Visibility,
				Bucket:     orignalsBucket,
				ObjectKey:  objectKey,
				Status:     ct.Pending,
				Variant:    ct.Original,
			})

			if err != nil {
				return ce.Wrap(
					ce.ErrInternal,
					err,
					"creating original file db entry error for file",
					input+": db: create file",
				).WithPublic("media service error")
			}

			for _, v := range variants {
				tele.Debug(ctx, "creating variants on db", "input", input, "variants", v)
				_, err := tx.CreateVariant(ctx, dbservice.File{
					Id:         fileId,
					Filename:   req.Filename,
					MimeType:   "image/webp",
					SizeBytes:  req.SizeBytes,
					Bucket:     variantsBucket,
					ObjectKey:  objectKey + "/" + v.String(),
					Visibility: req.Visibility,
					Status:     ct.Pending,
					Variant:    v,
				})
				if err != nil {
					return ce.Wrap(
						ce.ErrInternal,
						err,
						fmt.Sprintf("failed to create variant %s", v.String()),
					).WithPublic("media service error")
				}
			}

			url, err = m.S3.GenerateUploadURL(ctx, orignalsBucket, objectKey, exp)
			if err != nil {
				return ce.Wrap(
					ce.ErrInternal,
					err,
					fmt.Sprintf(
						"failed to create upload url for file with id %v:",
						fileId),
				).WithPublic("media service error")
			}
			return nil
		},
	)

	if errTx != nil {
		return 0, "", ce.Wrap(nil, errTx, input)
	}
	return fileId, url.String(), nil
}

// Returns an image download URL for the requested imageId and Variant.
// If the variant is not available it falls back to the original file.
func (m *MediaService) GetPublicImage(
	ctx context.Context,
	imgId ct.Id,
	variant ct.FileVariant,
) (string, error) {

	input := fmt.Sprintf("id: %d variant: %s", imgId, variant)

	if err := ct.ValidateBatch(imgId, variant); err != nil {
		return "", ce.Wrap(ce.ErrInvalidArgument, err, input)
	}

	var fm dbservice.File

	err := m.txRunner.RunTx(ctx, func(tx *dbservice.Queries) error {
		var err error

		if variant == ct.Original {
			fm, err = tx.GetFileById(ctx, imgId)
		} else {
			fm, err = tx.GetVariant(ctx, imgId, variant)
			if errors.Is(err, sql.ErrNoRows) || fm.Status == "pending" || fm.Status == "processing" {
				fm, err = tx.GetFileById(ctx, imgId)
			}
		}

		if err != nil {
			return mapDBError(err)
		}

		return validateFileStatus(fm)
	})

	if err != nil {
		return "", ce.Wrap(nil, err, input)
	}

	if fm.Visibility != ct.Public {
		return "", ce.New(ce.ErrPermissionDenied, ErrPermissionDenied, input).WithPublic("you don't have permission to view this image")
	}

	u, err := m.S3.GenerateDownloadURL(
		ctx, fm.Bucket, fm.ObjectKey, setExp(fm.Visibility),
	)
	if err != nil {
		return "", ce.Wrap(ce.ErrInternal, err, input+": s3: generate url")
	}

	return u.String(), nil
}

type FailedId struct {
	Id     ct.Id
	Status ct.UploadStatus
}

// Returns a id to download url pairs for
// an array of file ids and the prefered variant.
// Precondition for returning a file is the variant requested to exist in the database.
// Variant is common for all ids. If a variant is present but not completed
// returns url for the original format.
// GetImages does not accept original variants in batch request
func (m *MediaService) GetImages(ctx context.Context,
	imgIds ct.Ids, variant ct.FileVariant,
) (downUrls map[ct.Id]string, failedIds []FailedId, err error) {

	errMsg := fmt.Sprintf("get images: ids: %v variant: %s", imgIds, variant)

	if err := ct.ValidateBatch(imgIds, variant); err != nil {
		return nil, nil, ce.Wrap(ce.ErrInvalidArgument, err, errMsg)
	}

	var missingVariants ct.Ids
	var fms []dbservice.File

	failedIds = []FailedId{}
	err = m.txRunner.RunTx(ctx, func(tx *dbservice.Queries) error {
		var err error
		fms, missingVariants, err = tx.GetVariants(ctx, imgIds.Unique(), variant)
		if err != nil {
			return mapDBError(err)
		}

		if len(missingVariants) != 0 {
			originals, err := tx.GetFiles(ctx, missingVariants)
			if err != nil {
				return mapDBError(err)
			}
			fms = append(fms, originals...)
		}
		return nil
	})

	if err != nil {
		return nil, nil, ce.Wrap(nil, err, errMsg+": tx error")
	}

	downUrls = make(map[ct.Id]string, len(fms))
	for _, fm := range fms {
		if err := validateFileStatus(fm); err != nil {
			failedIds = append(failedIds, FailedId{Id: fm.Id, Status: fm.Status})
			tele.Warn(ctx,
				"failed to validate file status. @1", "error", err.Error())
			continue
		}
		url, err := m.S3.GenerateDownloadURL(ctx,
			fm.Bucket, fm.ObjectKey, setExp(fm.Visibility))
		if err != nil {
			return nil, nil, ce.Wrap(ce.ErrInternal, err, errMsg+": s3: generate url")
		}
		downUrls[fm.Id] = url.String()
	}
	return downUrls, failedIds, nil
}

// This is a call to validate an already uploaded file.
// Unvalidated files expire in 24 hours and are automatically
// deleted from file service.
// DEPRACATED
func (m *MediaService) ValidateUpload(ctx context.Context,
	fileId ct.Id, returnURL bool) (url string, err error) {

	input := fmt.Sprintf("file id: %d", fileId)

	if err := fileId.Validate(); err != nil {
		return url, ce.Wrap(ce.ErrInvalidArgument, err, input)
	}

	fileMeta, err := m.Queries.GetFileById(ctx, fileId)
	if err != nil {
		return "", ce.Wrap(nil, mapDBError(err), input)
	}

	if fileMeta.Status == ct.Failed {
		return url, ce.New(ce.ErrNotFound, ErrFailed, input).WithPublic("invalid file")
	}

	if fileMeta.Status != ct.Complete {
		if s3Err := m.S3.ValidateUpload(ctx,
			mapping.DbToModel(fileMeta)); s3Err != nil {
			if errors.Is(s3Err, ce.ErrInternal) {
				return "", ce.Wrap(nil, s3Err, input)
			}

			if err := m.S3.DeleteFile(ctx, fileMeta.Bucket, fileMeta.ObjectKey); err != nil {
				return url, ce.Wrap(nil, errors.Join(s3Err, err), input+": db: delete file")
			}
			if err := m.Queries.UpdateFileStatus(ctx, fileId, ct.Failed); err != nil {
				return url, ce.Wrap(nil, errors.Join(s3Err, err), input+": db: update file status")
			}
			return url, ce.Wrap(nil, s3Err, input).WithPublic("invalid file")
		}

		if err := m.Queries.UpdateFileStatus(ctx, fileId, ct.Complete); err != nil {
			tele.Error(ctx, "Media Service: Failed to update file status after file validation for @1 @2", "FileId", fileId, "error", err.Error())
			return url, ce.Wrap(ce.ErrInternal,
				fmt.Errorf("failed to update file status after file validation %w", err),
				input+": db: update file status",
			)
		}
		tele.Info(ctx, "Media Service: @1 successfully validated and marked as Complete", "FileId", fileId)
	}

	if returnURL {
		u, err := m.S3.GenerateDownloadURL(ctx,
			fileMeta.Bucket,
			fileMeta.ObjectKey,
			time.Duration(3*time.Minute),
		)
		if err != nil {
			tele.Info(ctx, "failed to fetch url for @1", "FileId", fileId)
			return "", nil
		}
		url = u.String()
	}
	return url, nil
}

// Sets 3 minutes expiration for private and 7 days exp for public
func setExp(v ct.FileVisibility) time.Duration {
	switch v {
	case ct.Private:
		return time.Duration(3 * time.Minute)
	case ct.Public:
		return time.Duration(7 * 24 * time.Hour)
	}
	return time.Duration(0)
}
