package application

import (
	"context"
	"errors"
	ce "social-network/shared/go/commonerrors"
	tele "social-network/shared/go/telemetry"
)

func (s *Application) RemoveImages(ctx context.Context, failedImages []int64) error {

	err := s.db.RemoveImages(ctx, failedImages)
	if err != nil {
		tele.Warn(ctx, "images @1 could not be deleted", "imageIds", failedImages)
	}

	return nil
}

func (s *Application) removeFailedImages(ctx context.Context, imagesToDelete []int64) {

	if len(imagesToDelete) > 0 {
		tele.Info(ctx, "removing image ids @1 from posts", "failedImageIds", imagesToDelete)
		err := s.RemoveImages(context.WithoutCancel(ctx), imagesToDelete)
		if err != nil {
			tele.Warn(ctx, "failed  to delete failed images @1 from posts: @2", "failedImageIds", imagesToDelete, "error", err.Error())
		}
	}
}

func (s *Application) removeFailedImagesAsync(ctx context.Context, imagesToDelete []int64) {
	go s.removeFailedImages(ctx, imagesToDelete)
}

func (s *Application) removeFailedImage(ctx context.Context, err error, imageId int64) {
	var commonError *ce.Error
	if errors.As(err, &commonError) {
		if err.(*ce.Error).IsClass(ce.ErrNotFound) {
			tele.Info(ctx, "removing image id @1 from posts", "failedImageId", imageId)
			err := s.RemoveImages(context.WithoutCancel(ctx), []int64{imageId})
			if err != nil {
				tele.Warn(ctx, "failed  to delete failed image @1 from posts: @2", "failedImageId", imageId, "error", err.Error())
			}
		}
	}
}

func (s *Application) removeFailedImageAsync(ctx context.Context, err error, imageId int64) {
	go s.removeFailedImage(ctx, err, imageId)
}
