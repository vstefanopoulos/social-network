package application

import (
	"context"
	"sync/atomic"
	"time"

	ct "social-network/shared/go/ct"
	tele "social-network/shared/go/telemetry"
)

var processingVariants atomic.Bool

// StartVariantWorker starts a background worker that periodically processes pending file variants
func (m *MediaService) StartVariantWorker(ctx context.Context, interval time.Duration) {
	tele.Info(ctx, "Initiating variant worker. @1", "interval", interval.String())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				if !processingVariants.CompareAndSwap(false, true) {
					continue
				}
				if err := m.processPendingVariants(ctx); err != nil {
					tele.Warn(ctx, "Error processing pending variants. @1", "error", err.Error())
				}
			case <-ctx.Done():
				tele.Info(ctx, "Variant worker stopped")
				return
			}
		}
	}()
}

// processPendingVariants queries for file_variants with status 'pending' and calls GenerateVariant for each
func (m *MediaService) processPendingVariants(ctx context.Context) error {
	defer processingVariants.Store(false)

	variants, err := m.Queries.GetPendingVariants(ctx)
	if err != nil {
		return err
	}

	for _, v := range variants {
		size, err := m.S3.GenerateVariant(ctx,
			v.SrcBucket,
			v.SrcObjectKey,
			v.Bucket,
			v.ObjectKey,
			v.Variant)
		if err != nil {
			tele.Warn(ctx, "Failed to generate variant for @1, @2", "id", v.Id, "variant", v.Variant, "error", err.Error())
			if updateErr := m.Queries.UpdateVariantStatusAndSize(ctx,
				v.Id,
				ct.Failed,
				size,
			); updateErr != nil {
				tele.Warn(ctx, "Failed to update status to failed. @1", "error", updateErr.Error())
			}
		} else {
			if updateErr := m.Queries.UpdateVariantStatusAndSize(ctx,
				v.Id,
				ct.Complete,
				size,
			); updateErr != nil {
				tele.Warn(ctx, "Failed to update status to complete. @1", "error", updateErr.Error())
			}
			tele.Info(ctx, "Successfully generated variant for @1  @2", "fileId", v.Id, "variant", v.Variant)
		}
	}

	return nil
}
