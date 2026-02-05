package application

import (
	"context"
	"errors"

	ce "social-network/shared/go/commonerrors"
	tele "social-network/shared/go/telemetry"
)

func checkPassword(storedPassword, newHashedPassword string) bool {

	tele.Debug(context.Background(), "Comparing passwords. @1 @2", "stored", storedPassword, "new", newHashedPassword)
	return storedPassword == newHashedPassword
}

func (s *Application) removeFailedImages(ctx context.Context, imagesToDelete []int64) {
	//==============pinpoint not found images============

	if len(imagesToDelete) > 0 {
		tele.Info(ctx, "removing avatar ids @1 from users", "failedImageIds", imagesToDelete)
		err := s.RemoveImages(context.WithoutCancel(ctx), imagesToDelete)
		if err != nil {
			tele.Warn(ctx, "failed  to delete failed images @1 from users: @2", "failedImageIds", imagesToDelete, "error", err.Error())
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

			tele.Info(ctx, "removing avatar id @1 from users", "failedImageId", imageId)
			err := s.RemoveImages(context.WithoutCancel(ctx), []int64{imageId})
			if err != nil {
				tele.Warn(ctx, "failed  to delete failed image @1 from users: @2", "failedImageId", imageId, "error", err.Error())
			}
		}

	}
}

func (s *Application) removeFailedImageAsync(ctx context.Context, err error, imageId int64) {
	go s.removeFailedImage(ctx, err, imageId)
}
