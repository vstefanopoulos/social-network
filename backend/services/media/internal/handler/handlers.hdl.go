package handler

import (
	"context"
	"time"

	"social-network/services/media/internal/application"
	pb "social-network/shared/gen-go/media"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/mapping"
	tele "social-network/shared/go/telemetry"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type MediaHandler struct {
	pb.UnimplementedMediaServiceServer
	Application *application.MediaService
	// Configs     configs.Server
}

// Provides image id and an upload URL that can only be accessed through container DNS.
// All uploads are marked with a false validation tag and must be validated through ValidateUpload handler.
// Unvalidated uploads expire after the defined `lifecycle.Expiration` on file services configuration.
//
// Usage:
//
//	exp := time.Duration(10 * time.Minute).Seconds()
//	var MediaService media.MediaServiceClient
//
//	mediaRes, err := MediaService.UploadImage(r.Context(), &media.UploadImageRequest{
//		Filename:   httpReq.AvatarName,
//		MimeType:   httpReq.AvatarType,
//		SizeBytes:  httpReq.AvatarSize,
//		Visibility: media.FileVisibility_PUBLIC,
//		Variants: []media.ImgVariant{
//			media.FileVariant_THUMBNAIL,
//			media.FileVariant_LARGE,
//		},
//		ExpirationSeconds: int64(exp),
//	})
func (m *MediaHandler) UploadImage(ctx context.Context,
	req *pb.UploadImageRequest) (*pb.UploadImageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request or file_meta is nil")
	}
	tele.Info(ctx, "upload image called @1", "request", req.String())

	// Convert variants
	variants := make([]ct.FileVariant, len(req.Variants))
	for i, v := range req.Variants {
		variants[i] = mapping.PbToCtFileVariant(v)
	}
	appReq := application.UploadImageReq{
		Filename:   req.Filename,
		MimeType:   req.MimeType,
		SizeBytes:  req.SizeBytes,
		Visibility: mapping.PbToCtFileVisibility(req.Visibility),
	}
	// Call application
	fileId, upUrl, err := m.Application.UploadImage(
		ctx,
		appReq,
		time.Duration(req.ExpirationSeconds)*time.Second,
		variants,
	)
	if err != nil {
		tele.Error(ctx, "failed to generate upload image url. @1 @2", "request:", req.String(), "error:", err.(*ce.Error).Error())
		return nil, ce.EncodeProto(err)
	}
	res := &pb.UploadImageResponse{
		FileId:    int64(fileId),
		UploadUrl: upUrl,
	}
	tele.Info(ctx, "upload image url generation success. @1 @2", "request", appReq, "response", res.String())
	return res, nil
}

// GetImage handles the gRPC request for retrieving an image download URL.
// Expiration time of link is set according to image visibility settings set on upload and
// is defined withing the methods of custom type 'FileVisibility'.
// Unvalidated uploads wont be fetched. In this case most likelly you will get a codes.NotFound error.
// If variant requested is not yet created the handler returns original
//
// Usage:
//
//	var MediaService media.MediaServiceClient
//	mediaRes, err := h.MediaService.GetImage(r.Context(), &media.GetImageRequest{
//		ImageId: 1,
//		Variant: media.FileVariant_ORIGINAL,
//	})
func (m *MediaHandler) GetImage(ctx context.Context,
	req *pb.GetImageRequest) (*pb.GetImageResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is nil")
	}
	tele.Info(ctx, "get image called @1", "request", req.String())

	// Call application
	downUrl, err := m.Application.GetPublicImage(ctx, ct.Id(req.ImageId), mapping.PbToCtFileVariant(req.Variant))
	if err != nil {
		tele.Error(ctx, "get image error", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}

	res := &pb.GetImageResponse{
		DownloadUrl: downUrl,
	}
	tele.Info(ctx, "get image success. @1 @2", "request", req.String(), "response", res.String())
	return res, nil
}

func (m *MediaHandler) GetImages(ctx context.Context,
	req *pb.GetImagesRequest) (*pb.GetImagesResponse, error) {
	if req == nil || req.ImgIds == nil {
		return nil, status.Error(codes.InvalidArgument, "request or img_ids is nil")
	}
	tele.Info(ctx, "get images called. @1", "request", req.String())
	// Convert img_ids to ct.Ids
	ids := make(ct.Ids, len(req.ImgIds.ImgIds))
	for i, id := range req.ImgIds.ImgIds {
		ids[i] = ct.Id(id)
	}

	// Call application
	downUrls, failedIds, err := m.Application.GetImages(ctx, ids, mapping.PbToCtFileVariant(req.Variant))
	if err != nil {
		tele.Error(ctx, "get images error", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}

	// Build response
	downloadUrls := make(map[int64]string, len(downUrls))
	for id, url := range downUrls {
		downloadUrls[int64(id)] = url
	}

	pbFailedIds := make([]*pb.FailedId, len(failedIds))
	for i, fId := range failedIds {
		pbFailedIds[i] = &pb.FailedId{
			FileId: int64(fId.Id),
			Status: mapping.CtToPbUploadStatus(fId.Status),
		}
	}
	res := &pb.GetImagesResponse{
		DownloadUrls: downloadUrls,
		FailedIds:    pbFailedIds,
	}
	tele.Info(ctx, "get images success. @1 @2", "request", req.String(), "response", res.String())
	return res, nil
}

// Checks if the upload matches the pre defined file metadata and configs FileService file constraints.
// Upon success all requested variants are generated, placed on file service and marked as completed on db rows.
// If validation fails the file cannot be retrived and will be deleted from file service after 24 hours
// As an exemption returns a FailedPrecondition upon ErrFailed return thus giving a chance for the file to be validated
// in the future.
func (m *MediaHandler) ValidateUpload(ctx context.Context,
	req *pb.ValidateUploadRequest) (*pb.ValidateUploadResponse, error) {
	if req == nil || req.FileId < 1 {
		return nil, status.Error(codes.InvalidArgument, "request or upload is nil")
	}

	tele.Info(ctx, "validate image called. @1", "request", req.String())

	// Call application
	url, err := m.Application.ValidateAndGenerateVariants(ctx, ct.Id(req.FileId), req.ReturnUrl)
	if err != nil {
		tele.Error(ctx, "validate image error", "request", req.String(), "error", err.Error())
		return nil, ce.EncodeProto(err)
	}
	res := &pb.ValidateUploadResponse{DownloadUrl: url}
	tele.Info(ctx, "validate image success. @1 @2", "request", req.String(), "response", res.String())
	return res, nil
}
