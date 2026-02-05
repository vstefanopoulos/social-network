package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/posts"
	"social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
	"time"
)

func (h *Handlers) createComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "createComment handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type CreateCommentJSONRequest struct {
			ParentId ct.Id          `json:"parent_id"`
			Body     ct.CommentBody `json:"comment_body"`

			ImageName string `json:"image_name"`
			ImageSize int64  `json:"image_size"`
			ImageType string `json:"image_type"`
		}

		httpReq := CreateCommentJSONRequest{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		if err := decoder.Decode(&httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "decoding error: "+err.Error())
			return
		}

		tele.Debug(ctx, "decoded create comment request. @1", "data", httpReq)

		if err := ct.ValidateStruct(httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "validation error: "+err.Error())
			return
		}

		imageVisibility := media.FileVisibility_PUBLIC
		//check post's visibility based on parent
		audience, err := h.PostsService.GetPostAudienceForComment(ctx, &posts.SimpleIdReq{Id: httpReq.ParentId.Int64()})
		if audience.Audience != "everyone" {
			imageVisibility = media.FileVisibility_PRIVATE
		}
		tele.Info(ctx, "comment audience is @1", "audience", audience.Audience)

		var ImageId ct.Id
		var uploadURL string
		if httpReq.ImageSize != 0 {
			exp := time.Duration(10 * time.Minute).Seconds()
			mediaRes, err := h.MediaService.UploadImage(ctx, &media.UploadImageRequest{
				Filename:          httpReq.ImageName,
				MimeType:          httpReq.ImageType,
				SizeBytes:         httpReq.ImageSize,
				Visibility:        imageVisibility,
				Variants:          []media.FileVariant{media.FileVariant_MEDIUM},
				ExpirationSeconds: int64(exp),
			})
			if err != nil {
				status, class := gorpc.Classify(err)
				utils.WriteJSON(ctx, w, status, class)
				return
			}
			ImageId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		grpcReq := posts.CreateCommentReq{
			CreatorId: int64(claims.UserId),
			ParentId:  httpReq.ParentId.Int64(),
			Body:      httpReq.Body.String(),
			ImageId:   ImageId.Int64(),
		}

		commentId, err := h.PostsService.CreateComment(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			// utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to create comment: %v", err.Error()))
			return
		}
		type httpResponse struct {
			CommentId ct.Id
			UserId    ct.Id
			FileId    ct.Id
			UploadUrl string
		}
		httpResp := httpResponse{
			CommentId: ct.Id(commentId.Id),
			UserId:    ct.Id(claims.UserId),
			FileId:    ImageId,
			UploadUrl: uploadURL}

		utils.WriteJSON(ctx, w, http.StatusOK, httpResp)
	}
}

func (h *Handlers) getCommentsByParentId() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getCommentsByParentId handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		// body, err := utils.JSON2Struct(&models.EntityIdPaginatedReq{}, r)
		// if err != nil {
		// 	utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
		// 	return
		// }

		v := r.URL.Query()
		entityId, err1 := utils.ParamGet(v, "entity_id", ct.Id(0), true)
		limit, err2 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err3 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2, err3); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.EntityIdPaginatedReq{
			RequesterId: claims.UserId,
			EntityId:    entityId.Int64(),
			Limit:       limit,
			Offset:      offset,
		}

		grpcResp, err := h.PostsService.GetCommentsByParentId(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to get comments for post id %v: %v: ", body.EntityId, err.Error()))
			return
		}

		tele.Debug(ctx, "retrieved comments. @1", "grpcResp", grpcResp)

		commentsResponse := []models.Comment{}
		for _, c := range grpcResp.Comments {
			comment := models.Comment{
				CommentId: ct.Id(c.CommentId),
				ParentId:  ct.Id(c.ParentId),
				Body:      ct.CommentBody(c.Body),
				User: models.User{
					UserId:    ct.Id(c.User.UserId),
					Username:  ct.Username(c.User.Username),
					AvatarId:  ct.Id(c.User.Avatar),
					AvatarURL: c.User.AvatarUrl,
				},
				ReactionsCount: int(c.ReactionsCount),
				CreatedAt:      ct.GenDateTime(c.CreatedAt.AsTime()),
				UpdatedAt:      ct.GenDateTime(c.UpdatedAt.AsTime()),
				LikedByUser:    c.LikedByUser,
				ImageId:        ct.Id(c.ImageId),
				ImageUrl:       c.ImageUrl,
			}
			commentsResponse = append(commentsResponse, comment)
		}

		err = utils.WriteJSON(ctx, w, http.StatusOK, commentsResponse)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to send comments for post %v : %v", entityId, err.Error()))
			return
		}
	}
}

func (h *Handlers) editComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		tele.Info(ctx, "editComment handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type EditCommentJSONRequest struct {
			CommentId   ct.Id
			Body        ct.CommentBody `json:"comment_body"`
			DeleteImage bool           `json:"delete_image"`

			ImageName string `json:"image_name"`
			ImageSize int64  `json:"image_size"`
			ImageType string `json:"image_type"`
		}

		httpReq := EditCommentJSONRequest{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		if err := decoder.Decode(&httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}
		var err error
		httpReq.CommentId, err = utils.PathValueGet(r, "comment_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		if err := ct.ValidateStruct(httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		var ImageId ct.Id
		var uploadURL string
		if httpReq.ImageSize != 0 {
			exp := time.Duration(10 * time.Minute).Seconds()
			mediaRes, err := h.MediaService.UploadImage(ctx, &media.UploadImageRequest{
				Filename:          httpReq.ImageName,
				MimeType:          httpReq.ImageType,
				SizeBytes:         httpReq.ImageSize,
				Visibility:        media.FileVisibility_PUBLIC,
				Variants:          []media.FileVariant{media.FileVariant_MEDIUM},
				ExpirationSeconds: int64(exp),
			})
			if err != nil {
				utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
				return
			}
			ImageId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		grpcReq := posts.EditCommentReq{
			CreatorId:   int64(claims.UserId),
			CommentId:   httpReq.CommentId.Int64(),
			Body:        httpReq.Body.String(),
			ImageId:     ImageId.Int64(),
			DeleteImage: httpReq.DeleteImage,
		}

		_, err = h.PostsService.EditComment(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to create comment: %v", err.Error()))
			return
		}
		type httpResponse struct {
			UserId    ct.Id
			FileId    ct.Id
			UploadUrl string
		}
		httpResp := httpResponse{
			UserId:    ct.Id(claims.UserId),
			FileId:    ImageId,
			UploadUrl: uploadURL}

		utils.WriteJSON(ctx, w, http.StatusOK, httpResp)
	}
}

func (h *Handlers) deleteComment() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "deleteComment handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		// body, err := utils.JSON2Struct(&models.GenericReq{}, r)
		// if err != nil {
		// 	utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
		// 	return
		// }
		body := &models.GenericReq{}
		var err error
		body.EntityId, err = utils.PathValueGet(r, "comment_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.GenericReq{
			RequesterId: int64(claims.UserId),
			EntityId:    body.EntityId.Int64(),
		}

		_, err = h.PostsService.DeleteComment(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to delete comment with id %v: %v", body.EntityId, err.Error()))
			return
		}

	}
}
