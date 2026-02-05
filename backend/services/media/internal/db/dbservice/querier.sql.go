package dbservice

import (
	"context"
	ct "social-network/shared/go/ct"
)

type Querier interface {
	CreateFile(ctx context.Context, fm File) (fileId ct.Id, err error)

	GetFileById(ctx context.Context, fileId ct.Id) (fm File, err error)

	GetFiles(
		ctx context.Context, ids ct.Ids,
	) ([]File, error)

	UpdateFileStatus(
		ctx context.Context,
		fileId ct.Id,
		status ct.UploadStatus,
	) error

	CreateVariant(ctx context.Context, fm File) (fileId ct.Id, err error)

	GetVariant(ctx context.Context, fileId ct.Id,
		variant ct.FileVariant) (fm File, err error)

	GetAllVariants(
		ctx context.Context,
		fileId ct.Id,
	) (fms []File, err error)

	GetVariants(
		ctx context.Context,
		fileIds ct.Ids,
		variant ct.FileVariant,
	) (fms []File, notComplete []ct.Id, err error)

	UpdateVariantStatusAndSize(
		ctx context.Context,
		varId ct.Id,
		status ct.UploadStatus,
		size int64,
	) error

	GetPendingVariants(
		ctx context.Context) (pending []Variant, err error)

	// StartStaleFilesWorker(ctx context.Context)
	MarkStaleFilesFailed(ctx context.Context) error
}

var _ Querier = (*Queries)(nil)
