package application

import (
	"database/sql"
	"errors"
	"fmt"
	"social-network/services/media/internal/db/dbservice"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"time"
)

var (
	ErrReqValidation    = errors.New("request validation error")               // invalid arguments
	ErrNotValidated     = errors.New("file not yet validated")                 // means that validation is pending
	ErrFailed           = errors.New("file has permanently failed validation") // means that file validation has failed permanently
	ErrNotFound         = errors.New("not found")                              // Usually equivalent to sql.ErrNoRows
	ErrInternal         = errors.New("internal error")
	ErrValidateStatus   = errors.New("validate status error")
	ErrPermissionDenied = errors.New("permission denied")
)

// Maps a file status to application errors and returns error.
// Caller decides if adding extra info about the file
func validateFileStatus(fm dbservice.File) error {
	errMsg := fmt.Sprintf(
		"file id %v file name %v status %v",
		fm.Id,
		fm.Filename,
		fm.Status,
	)

	if fm.Status == ct.Complete {
		return nil
	}

	if err := fm.Status.Validate(); err != nil {
		return ce.Wrap(ce.ErrDataLoss, err, errMsg)
	}

	if fm.Status == ct.Failed {
		return ce.Wrap(ce.ErrNotFound, ErrValidateStatus, errMsg).
			WithPublic("file permenantly failed")
	}

	if fm.Status == ct.Pending || fm.Status == ct.Processing {
		// TODO: Think if I should validate here
		return ce.Wrap(ce.ErrFailedPrecondition, ErrValidateStatus, errMsg).
			WithPublic("file not yet validated")
	}

	return nil
}

// Maps an db error to application custom error types
// sql.NoRows == ErrNotFound all other errors ErrInternal.
func mapDBError(err error) error {
	if err == nil {
		return nil
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ce.New(ce.ErrNotFound, err).WithPublic("not found")
	}

	return ce.New(ce.ErrInternal, err).WithPublic("internal media error")
}

// validateUploadRequest validates all inputs required to create an image upload.
// It ensures the request metadata is well-formed, allowed by configuration,
// and safe to process.
//
// Validation rules:
//   - Filename must be non-empty.
//   - MimeType must be provided and allowed by file service constraints.
//   - SizeBytes must be greater than zero.
//   - Visibility must be a valid enum value.
//   - Expiration must be between 1 minute and 24 hours.
//   - At least one file variant must be provided.
//   - Each variant must be valid.
//   - The ct.Original variant is not allowed, as it is created implicitly.
//
// Returns a descriptive error on validation failure, or nil if the request is valid.
func (m *MediaService) validateUploadRequest(
	req UploadImageReq,
	exp time.Duration,
	variants []ct.FileVariant,
) error {

	if req.Filename == "" {
		return fmt.Errorf("upload image: invalid filename %q", req.Filename)
	}

	if req.MimeType == "" {
		return fmt.Errorf("upload image: missing mime type for file %v", req.Filename)
	}

	if !m.Cfgs.FileService.FileConstraints.AllowedMIMEs[req.MimeType] {
		return fmt.Errorf("upload image: mime type %q not allowed  for file %v", req.MimeType, req.Filename)
	}

	if req.SizeBytes < 1 || req.SizeBytes > m.Cfgs.FileService.FileConstraints.MaxImageUpload {
		return fmt.Errorf("upload image: invalid size %d for file %v", req.SizeBytes, req.Filename)
	}

	if err := req.Visibility.Validate(); err != nil {
		return fmt.Errorf("upload image: invalid visibility %v for file %v", req.Visibility, req.Filename)
	}

	if exp < time.Minute || exp > 24*time.Hour {
		return fmt.Errorf("upload image: invalid expiration %v for file %v", exp, req.Filename)
	}

	for _, v := range variants {
		if err := v.Validate(); err != nil {
			return fmt.Errorf("invalid variant %v for file %v", v, req.Filename)
		}
		if v == ct.Original {
			return fmt.Errorf("original is not a variant %v for file %v", v, req.Filename)
		}
	}

	return nil
}
