package retrievemedia

import (
	"context"
	"fmt"
	"maps"

	"social-network/shared/gen-go/media"
	ce "social-network/shared/go/commonerrors"
	"social-network/shared/go/ct"
	"social-network/shared/go/mapping"
	tele "social-network/shared/go/telemetry"
)

// GetImages returns a map[imageId]imageUrl, using cache + batch RPC.
func (h *MediaRetriever) GetImages(ctx context.Context, imageIds ct.Ids, variant media.FileVariant) (map[int64]string, []int64, error) {
	input := fmt.Sprintf("ids %v, variant: %v", imageIds, variant)

	if len(imageIds) == 0 {
		tele.Warn(ctx, "get media called with empty ids slice")
		return nil, nil, nil
	}

	if err := imageIds.Validate(); err != nil {
		return nil, nil, ce.New(ce.ErrInvalidArgument, err, input)
	}

	uniqueImageIds := imageIds.Unique()
	images := make(map[int64]string, len(uniqueImageIds))
	var missingImages ct.Ids
	var imagesToDelete []int64

	ctVariant := mapping.PbToCtFileVariant(variant)
	if err := ctVariant.Validate(); err != nil {
		// If variant is invalid, we probably can't do anything meaningful.
		// Returning error or empty map? Returning error seems safest.
		return nil, nil, ce.New(ce.ErrInvalidArgument, err, input)
	}

	// Redis lookup for images
	for _, imageId := range uniqueImageIds {
		key, err := ct.ImageKey{Id: imageId, Variant: ctVariant}.GenKey()
		if err != nil {
			tele.Warn(ctx, "failed to construct redis key for image @1: @2", "imageId", imageId, "error", err.Error())
			missingImages = append(missingImages, imageId)
			continue
		}

		imageURL, err := h.cache.GetStr(ctx, key)
		if err == nil {
			tele.Info(ctx, "Got url for image @1 from redis", "imageId", imageId)
			images[imageId.Int64()] = imageURL.(string)
		} else {
			missingImages = append(missingImages, imageId)
		}
	}

	// Batch RPC for missing images
	if len(missingImages) > 0 {
		// fmt.Println("calling media for these images", missingImages)
		req := &media.GetImagesRequest{
			ImgIds:  &media.ImageIds{ImgIds: missingImages.Int64()},
			Variant: variant,
		}

		resp, err := h.client.GetImages(ctx, req)
		if err != nil {
			return nil, nil, ce.DecodeProto(err, input)
		}

		// merge with redis map
		maps.Copy(images, resp.DownloadUrls)

		// Cache the new results
		for id, url := range resp.DownloadUrls {
			key, err := ct.ImageKey{Id: ct.Id(id), Variant: ctVariant}.GenKey()
			if err == nil {
				_ = h.cache.SetStr(ctx, key, url, h.ttl)
			} else {
				tele.Warn(ctx, "failed to construct redis key for image @1: @2", "imageId", id, "error", err.Error())
			}
		}

		//==============pinpoint not found images============

		// Build set of all requested IDs
		requested := make(map[int64]struct{}, len(uniqueImageIds))
		for _, imageId := range uniqueImageIds {
			requested[imageId.Int64()] = struct{}{}
		}

		// Remove images found in redis and from media service
		for id := range images {
			delete(requested, id)
		}

		// Remove images in failedIds unless status failed
		for _, failed := range resp.FailedIds {
			if failed.GetStatus() != media.UploadStatus_UPLOAD_STATUS_FAILED {
				delete(requested, failed.FileId)
			}
		}
		//now requested only contains failed and not found anywhere
		for id := range requested {
			imagesToDelete = append(imagesToDelete, id)
		}
	}

	return images, imagesToDelete, nil
}

// func toCtVariant(v media.FileVariant) (ct.FileVariant, error) {
// 	switch v {
// 	case media.FileVariant_THUMBNAIL:
// 		return ct.ImgThumbnail, nil
// 	case media.FileVariant_SMALL:
// 		return ct.ImgSmall, nil
// 	case media.FileVariant_MEDIUM:
// 		return ct.ImgMedium, nil
// 	case media.FileVariant_LARGE:
// 		return ct.ImgLarge, nil
// 	case media.FileVariant_ORIGINAL:
// 		return ct.Original, nil
// 	default:
// 		return "", fmt.Errorf("unknown media variant: %v", v)
// 	}
// }

// GetImage returns a single image url, using cache + batch RPC.
func (h *MediaRetriever) GetImage(ctx context.Context, imageId int64, variant media.FileVariant) (string, *ce.Error) {
	input := fmt.Sprintf("id %v, variant: %v", imageId, variant)

	if err := ct.Id(imageId).Validate(); err != nil {
		return "", ce.New(ce.ErrInvalidArgument, err, input)
	}

	ctVariant := mapping.PbToCtFileVariant(variant)
	if err := ctVariant.Validate(); err != nil {
		return "", ce.New(ce.ErrInvalidArgument, err, input)
	}

	// Redis lookup for images
	key, err := ct.ImageKey{Id: ct.Id(imageId), Variant: ctVariant}.GenKey()
	if err != nil {
		tele.Warn(ctx, "failed to construct redis key for image @1: @1", "imageId", imageId, "error", err.Error())
	}

	imageURL, err := h.cache.GetStr(ctx, key)
	if err == nil {
		tele.Info(ctx, "Got image url for image @1 from redis", "imageId", imageId)
		return imageURL.(string), nil
	}

	resp, err := h.client.GetImage(ctx, &media.GetImageRequest{
		ImageId: imageId,
		Variant: variant,
	})
	if err != nil {
		return "", ce.DecodeProto(err, input)
	}

	//if err, check if need to delete image

	// Cache the new result
	key, err = ct.ImageKey{Id: ct.Id(imageId), Variant: ctVariant}.GenKey()
	if err == nil {
		_ = h.cache.SetStr(ctx, key, resp.DownloadUrl, h.ttl)
	} else {
		tele.Warn(ctx, "failed to construct redis key for image @1: @2", "imageId", imageId, "error", err.Error())
	}

	return resp.DownloadUrl, nil
}
