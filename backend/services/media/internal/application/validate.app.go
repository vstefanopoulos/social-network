package application

import (
	"context"
	"errors"
	"fmt"
	"social-network/services/media/internal/client"
	"social-network/services/media/internal/db/dbservice"
	"social-network/services/media/internal/mapping"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"
	"time"
)

func (m *MediaService) ValidateAndGenerateVariants(ctx context.Context,
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

	// Leave early and let worker handle the variants
	if fileMeta.Status == ct.Complete {
		return m.urlOption(ctx, fileMeta, returnURL)
	}

	variants, err := m.getAllVariants(ctx, fileId)
	if err != nil {
		return url, ce.Wrap(nil, err, input)
	}

	Err := m.S3.ValidateAndCreateVariants(ctx,
		mapping.DbToModel(fileMeta),
		variants,
	)
	if Err != nil {
		if Err.IsClass(ce.ErrInternal) {
			return "", ce.Wrap(nil, Err, input)
		}
		return url, ce.Wrap(
			nil,
			errors.Join(
				Err,
				m.deleteFailedFile(ctx, fileId, fileMeta), // join possible internal errors from db
			),
			input,
		).WithPublic("invalid file")
	}

	// Update database with new statuses
	if err := m.markStatusComplete(ctx, fileId, variants); err != nil {
		return "", ce.Wrap(nil, err, input)
	}

	tele.Info(ctx, "Media Service: @1 successfully validated and marked as Complete", "FileId", fileId)

	return m.urlOption(ctx, fileMeta, returnURL)
}

// Gets all file variant rows linked to fileId with status other than complete and converts to
// client.VariantsToGenerate slice.
func (m *MediaService) getAllVariants(
	ctx context.Context,
	fileId ct.Id,
) ([]client.VariantToGenerate, error) {

	variants, err := m.Queries.GetAllVariants(ctx, fileId)
	if err != nil {
		return nil, ce.Wrap(ce.ErrInternal, err)
	}
	tele.Debug(ctx, "variants to create", "vars", variants)

	var vars []client.VariantToGenerate
	for _, v := range variants {
		tele.Debug(ctx, "variants to create", "var", v)
		vars = append(vars, mapClVariantToDbServiceVariant(v))
	}
	return vars, nil
}

// Converts dbservice.File to client.VariantToGenerate.
func mapClVariantToDbServiceVariant(v dbservice.File) client.VariantToGenerate {
	return client.VariantToGenerate{
		Id:      v.Id,
		Bucket:  v.Bucket,
		ObjKey:  v.ObjectKey,
		Variant: v.Variant,
	}
}

// Deletes file from s3 service and marks its status as failed on db row.
func (m *MediaService) deleteFailedFile(
	ctx context.Context,
	fileId ct.Id,
	fm dbservice.File,
) *ce.Error {

	if err := m.S3.DeleteFile(ctx, fm.Bucket, fm.ObjectKey); err != nil {
		return ce.Wrap(nil, err)
	}

	if err := m.Queries.UpdateFileStatus(ctx, fileId, ct.Failed); err != nil {
		return ce.Wrap(nil, err)
	}
	return nil

}

// Marks status on db rows of file and variants as completed.
func (m *MediaService) markStatusComplete(
	ctx context.Context,
	fileID ct.Id,
	vars []client.VariantToGenerate,
) error {
	return m.txRunner.RunTx(ctx, func(tx *dbservice.Queries) error {
		// Update file
		if err := tx.UpdateFileStatus(ctx, fileID, ct.Complete); err != nil {
			tele.Error(
				ctx,
				"Media Service: Failed to update file status after file validation",
				"FileId", fileID,
				"error", err.Error(),
			)
			return ce.Wrap(
				ce.ErrInternal,
				fmt.Errorf("failed to update file status after file validation %w", err),
			)
		}

		ids := make([]ct.Id, 0, len(vars))
		sizes := make([]int64, 0, len(vars))

		for _, v := range vars {
			ids = append(ids, v.Id)
			sizes = append(sizes, v.Size)
		}

		if err := tx.UpdateVariantsStatusAndSize(ctx, ids, ct.Complete, sizes); err != nil {
			return ce.Wrap(
				ce.ErrInternal,
				fmt.Errorf("failed to update variant statuses: %w", err),
			)
		}

		tele.Debug(ctx, "updated variant statuses to complete", "count", len(vars))

		// Update variants
		// for _, v := range vars {
		// 	// Todo: Make this more efficient with an atomic query
		// 	if err := tx.UpdateVariantStatusAndSize(ctx, v.Id, ct.Complete, v.Size); err != nil {
		// 		return ce.Wrap(
		// 			ce.ErrInternal,
		// 			fmt.Errorf("%w: failed to update variant status: %v", err, v),
		// 		)
		// 	}
		// 	tele.Debug(ctx, "updated variant status to complete", "variant", v.ObjKey)
		// }

		return nil
	})
}

// Helper for optional url return.
func (m *MediaService) urlOption(ctx context.Context, fm dbservice.File, do bool) (string, error) {
	if !do {
		return "", nil
	}
	u, err := m.S3.GenerateDownloadURL(ctx,
		fm.Bucket,
		fm.ObjectKey,
		time.Duration(1*time.Minute),
	)
	if err != nil {
		tele.Info(ctx, "failed to fetch url for @1", "FileId", fm.Id)
		return "", nil
	}
	return u.String(), nil
}
