package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/shared/gen-go/common"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/posts"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
	"time"
)

func (h *Handlers) getPostById() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getPostById handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}
		postId, err := utils.PathValueGet(r, "post_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.GenericReq{
			RequesterId: int64(claims.UserId),
			EntityId:    postId.Int64(),
		}

		grpcResp, err := h.PostsService.GetPostById(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to get post with id %v: %s", body.EntityId, err.Error()))
			return
		}

		tele.Info(ctx, "retrieved post by id. @1", "grpcResp", grpcResp)

		selectedAudience := make([]models.User, 0, len(grpcResp.SelectedAudienceUsers.Users))

		for _, u := range grpcResp.SelectedAudienceUsers.Users {

			selectedAudience = append(selectedAudience, models.User{
				UserId:    ct.Id(u.UserId),
				Username:  ct.Username(u.Username),
				AvatarId:  ct.Id(u.Avatar),
				AvatarURL: u.AvatarUrl,
			})
		}

		post := models.Post{
			PostId: ct.Id(grpcResp.PostId),
			Body:   ct.PostBody(grpcResp.PostBody),
			User: models.User{
				UserId:    ct.Id(grpcResp.User.UserId),
				Username:  ct.Username(grpcResp.User.Username),
				AvatarId:  ct.Id(grpcResp.User.Avatar),
				AvatarURL: grpcResp.User.AvatarUrl,
			},
			GroupId:               ct.Id(grpcResp.GroupId),
			Audience:              ct.Audience(grpcResp.Audience),
			CommentsCount:         int(grpcResp.CommentsCount),
			ReactionsCount:        int(grpcResp.ReactionsCount),
			LastCommentedAt:       ct.GenDateTime(grpcResp.LastCommentedAt.AsTime()),
			CreatedAt:             ct.GenDateTime(grpcResp.CreatedAt.AsTime()),
			UpdatedAt:             ct.GenDateTime(grpcResp.UpdatedAt.AsTime()),
			LikedByUser:           grpcResp.LikedByUser,
			ImageId:               ct.Id(grpcResp.ImageId),
			ImageUrl:              grpcResp.ImageUrl,
			SelectedAudienceUsers: selectedAudience,
		}

		err = utils.WriteJSON(ctx, w, http.StatusOK, post)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send post by id")
			return
		}

	}
}

func (h *Handlers) getMostPopularPostInGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getMostPopularPostInGroup handler called")

		groupId, err := utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.SimpleIdReq{
			Id: groupId.Int64(),
		}

		grpcResp, err := h.PostsService.GetMostPopularPostInGroup(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to get post with id %v: %s", body.EntityId, err.Error()))
			return
		}

		tele.Info(ctx, "retrieved most popular post in group @1", "grpcResp", grpcResp)

		var post *models.Post

		if grpcResp.PostId != 0 {

			post = &models.Post{
				PostId: ct.Id(grpcResp.PostId),
				Body:   ct.PostBody(grpcResp.PostBody),
				User: models.User{
					UserId:    ct.Id(grpcResp.User.UserId),
					Username:  ct.Username(grpcResp.User.Username),
					AvatarId:  ct.Id(grpcResp.User.Avatar),
					AvatarURL: grpcResp.User.AvatarUrl,
				},
				GroupId:         ct.Id(grpcResp.GroupId),
				Audience:        ct.Audience(grpcResp.Audience),
				CommentsCount:   int(grpcResp.CommentsCount),
				ReactionsCount:  int(grpcResp.ReactionsCount),
				LastCommentedAt: ct.GenDateTime(grpcResp.LastCommentedAt.AsTime()),
				CreatedAt:       ct.GenDateTime(grpcResp.CreatedAt.AsTime()),
				UpdatedAt:       ct.GenDateTime(grpcResp.UpdatedAt.AsTime()),
				LikedByUser:     grpcResp.LikedByUser,
				ImageId:         ct.Id(grpcResp.ImageId),
				ImageUrl:        grpcResp.ImageUrl,
			}
		}

		err = utils.WriteJSON(ctx, w, http.StatusOK, post)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send most popular post in group")
			return
		}

	}
}

func (h *Handlers) createPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "createPost handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type CreatePostJSONRequest struct {
			Body        ct.PostBody `json:"post_body"`
			GroupId     ct.Id       `json:"group_id" validate:"nullable"`
			Audience    ct.Audience `json:"audience"`
			AudienceIds ct.Ids      `json:"audience_ids" validate:"nullable"`

			ImageName string `json:"image_name"`
			ImageSize int64  `json:"image_size"`
			ImageType string `json:"image_type"`
		}

		httpReq := CreatePostJSONRequest{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		if err := decoder.Decode(&httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		if err := ct.ValidateStruct(httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		imageVisibility := media.FileVisibility_PUBLIC
		if httpReq.Audience.String() != "everyone" {
			imageVisibility = media.FileVisibility_PRIVATE
		}

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
				utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
				return
			}
			ImageId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		grpcReq := posts.CreatePostReq{
			CreatorId: int64(claims.UserId),
			Body:      httpReq.Body.String(),
			GroupId:   httpReq.GroupId.Int64(),
			Audience:  httpReq.Audience.String(),
			AudienceIds: &common.UserIds{
				Values: httpReq.AudienceIds.Int64(),
			},
			ImageId: ImageId.Int64(),
		}

		postId, err := h.PostsService.CreatePost(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to create post: %v", err.Error()))
			return
		}
		type httpResponse struct {
			PostId    ct.Id
			UserId    ct.Id
			FileId    ct.Id
			UploadUrl string
		}
		httpResp := httpResponse{
			PostId:    ct.Id(postId.Id),
			UserId:    ct.Id(claims.UserId),
			FileId:    ImageId,
			UploadUrl: uploadURL,
		}
		tele.Info(ctx, "created post successfully")
		utils.WriteJSON(ctx, w, http.StatusOK, httpResp)
	}
}

func (h *Handlers) editPost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "editPost handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type EditPostJSONRequest struct {
			PostId      ct.Id
			NewBody     ct.PostBody `json:"post_body"`
			Audience    ct.Audience `json:"audience"`
			AudienceIds ct.Ids      `json:"audience_ids" validate:"nullable"`
			DeleteImage bool        `json:"delete_image"`

			ImageName string `json:"image_name"`
			ImageSize int64  `json:"image_size"`
			ImageType string `json:"image_type"`
		}

		httpReq := EditPostJSONRequest{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		if err := decoder.Decode(&httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		var err error
		httpReq.PostId, err = utils.PathValueGet(r, "post_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		if err := ct.ValidateStruct(httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		imageVisibility := media.FileVisibility_PUBLIC
		if httpReq.Audience.String() != "everyone" {
			imageVisibility = media.FileVisibility_PRIVATE
		}

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
				utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
				return
			}
			ImageId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		grpcReq := posts.EditPostReq{
			RequesterId: int64(claims.UserId),
			PostId:      int64(httpReq.PostId),
			Body:        httpReq.NewBody.String(),
			Audience:    httpReq.Audience.String(),
			AudienceIds: &common.UserIds{
				Values: httpReq.AudienceIds.Int64(),
			},
			ImageId:     ImageId.Int64(),
			DeleteImage: httpReq.DeleteImage,
		}

		_, err = h.PostsService.EditPost(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to create post: %v", err.Error()))
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

func (h *Handlers) deletePost() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "deletePost handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		postId, err := utils.PathValueGet(r, "post_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.GenericReq{
			RequesterId: int64(claims.UserId),
			EntityId:    postId.Int64(),
		}

		_, err = h.PostsService.DeletePost(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to delete post with id %v: %v", body.EntityId, err.Error()))
			return
		}

	}
}
