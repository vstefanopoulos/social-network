package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"social-network/shared/gen-go/posts"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
)

func (h *Handlers) getPublicFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getPublicFeed handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1) //TODO remove all these panics
		}

		v := r.URL.Query()

		limit, err1 := utils.ParamGet(v, "limit", int32(1), false)

		offset, err2 := utils.ParamGet(v, "offset", int32(0), false)

		if err := errors.Join(err1, err2); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.GenericPaginatedReq{
			RequesterId: claims.UserId,
			Limit:       limit,
			Offset:      offset,
		}

		grpcResp, err := h.PostsService.GetPublicFeed(ctx, &grpcReq)

		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to get public feed: "+err.Error())
			return
		}

		tele.Info(ctx, "retrieved public feed. @1", "grpcResp", grpcResp)

		postsResponse := []models.Post{}
		for _, p := range grpcResp.Posts {
			post := models.Post{
				PostId: ct.Id(p.PostId),
				Body:   ct.PostBody(p.PostBody),
				User: models.User{
					UserId:    ct.Id(p.User.UserId),
					Username:  ct.Username(p.User.Username),
					AvatarId:  ct.Id(p.User.Avatar),
					AvatarURL: p.User.AvatarUrl,
				},
				GroupId:         ct.Id(p.GroupId),
				Audience:        ct.Audience(p.Audience),
				CommentsCount:   int(p.CommentsCount),
				ReactionsCount:  int(p.ReactionsCount),
				LastCommentedAt: ct.GenDateTime(p.LastCommentedAt.AsTime()),
				CreatedAt:       ct.GenDateTime(p.CreatedAt.AsTime()),
				UpdatedAt:       ct.GenDateTime(p.UpdatedAt.AsTime()),
				LikedByUser:     p.LikedByUser,
				ImageId:         ct.Id(p.ImageId),
				ImageUrl:        p.ImageUrl,
			}
			postsResponse = append(postsResponse, post)
		}

		err = utils.WriteJSON(ctx, w, http.StatusOK, postsResponse)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send public feed")
			return
		}

	}
}

func (h *Handlers) getPersonalizedFeed() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getPersonalizedFeed handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)

		if !ok {
			tele.Error(ctx, "problem fetching claims")
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "problem with jwt claims")
			return
		}

		v := r.URL.Query()

		limit, err1 := utils.ParamGet(v, "limit", int32(1), false)

		offset, err2 := utils.ParamGet(v, "offset", int32(0), false)

		if err := errors.Join(err1, err2); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.GetPersonalizedFeedReq{
			RequesterId: claims.UserId,
			Limit:       limit,
			Offset:      offset,
		}

		grpcResp, err := h.PostsService.GetPersonalizedFeed(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to get personalized feed: "+err.Error())
			return
		}

		tele.Info(ctx, "retrieved personalized feed. @1", "grpcResp", grpcResp)

		postsResponse := []models.Post{}
		for _, p := range grpcResp.Posts {
			post := models.Post{
				PostId: ct.Id(p.PostId),
				Body:   ct.PostBody(p.PostBody),
				User: models.User{
					UserId:    ct.Id(p.User.UserId),
					Username:  ct.Username(p.User.Username),
					AvatarId:  ct.Id(p.User.Avatar),
					AvatarURL: p.User.AvatarUrl,
				},
				GroupId:         ct.Id(p.GroupId),
				Audience:        ct.Audience(p.Audience),
				CommentsCount:   int(p.CommentsCount),
				ReactionsCount:  int(p.ReactionsCount),
				LastCommentedAt: ct.GenDateTime(p.LastCommentedAt.AsTime()),
				CreatedAt:       ct.GenDateTime(p.CreatedAt.AsTime()),
				UpdatedAt:       ct.GenDateTime(p.UpdatedAt.AsTime()),
				LikedByUser:     p.LikedByUser,
				ImageId:         ct.Id(p.ImageId),
				ImageUrl:        p.ImageUrl,
			}
			postsResponse = append(postsResponse, post)
		}

		err = utils.WriteJSON(ctx, w, http.StatusOK, postsResponse)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send public feed")
			return
		}

	}
}

func (h *Handlers) getUserPostsPaginated() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getUserPostsPaginated handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		v := r.URL.Query()
		creatorId, err1 := utils.PathValueGet(r, "user_id", ct.Id(0), true)
		limit, err2 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err3 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2, err3); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.GetUserPostsReq{
			RequesterId: claims.UserId,
			CreatorId:   creatorId.Int64(),
			Limit:       limit,
			Offset:      offset,
		}

		grpcResp, err := h.PostsService.GetUserPostsPaginated(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to get personalized feed: "+err.Error())
			return
		}

		tele.Info(ctx, "retrieved personalized feed. @1", "grpcResp", grpcResp)

		postsResponse := []models.Post{}
		for _, p := range grpcResp.Posts {
			post := models.Post{
				PostId: ct.Id(p.PostId),
				Body:   ct.PostBody(p.PostBody),
				User: models.User{
					UserId:    ct.Id(p.User.UserId),
					Username:  ct.Username(p.User.Username),
					AvatarId:  ct.Id(p.User.Avatar),
					AvatarURL: p.User.AvatarUrl,
				},
				GroupId:         ct.Id(p.GroupId),
				Audience:        ct.Audience(p.Audience),
				CommentsCount:   int(p.CommentsCount),
				ReactionsCount:  int(p.ReactionsCount),
				LastCommentedAt: ct.GenDateTime(p.LastCommentedAt.AsTime()),
				CreatedAt:       ct.GenDateTime(p.CreatedAt.AsTime()),
				UpdatedAt:       ct.GenDateTime(p.UpdatedAt.AsTime()),
				LikedByUser:     p.LikedByUser,
				ImageId:         ct.Id(p.ImageId),
				ImageUrl:        p.ImageUrl,
			}
			postsResponse = append(postsResponse, post)
		}

		err = utils.WriteJSON(ctx, w, http.StatusOK, postsResponse)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to send user %v posts: %v", creatorId, err.Error()))
			return
		}

	}
}

func (h *Handlers) getGroupPostsPaginated() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getGroupPostsPaginated handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		v := r.URL.Query()
		groupId, err1 := utils.PathValueGet(r, "group_id", ct.Id(0), true)
		limit, err2 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err3 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2, err3); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := posts.GetGroupPostsReq{
			RequesterId: claims.UserId,
			GroupId:     groupId.Int64(),
			Limit:       limit,
			Offset:      offset,
		}

		grpcResp, err := h.PostsService.GetGroupPostsPaginated(ctx, &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to get group feed: "+err.Error())
			return
		}

		tele.Info(ctx, "retrieved group feed. @1", "grpcResp", grpcResp)

		postsResponse := []models.Post{}
		for _, p := range grpcResp.Posts {
			post := models.Post{
				PostId: ct.Id(p.PostId),
				Body:   ct.PostBody(p.PostBody),
				User: models.User{
					UserId:    ct.Id(p.User.UserId),
					Username:  ct.Username(p.User.Username),
					AvatarId:  ct.Id(p.User.Avatar),
					AvatarURL: p.User.AvatarUrl,
				},
				GroupId:         ct.Id(p.GroupId),
				Audience:        ct.Audience(p.Audience),
				CommentsCount:   int(p.CommentsCount),
				ReactionsCount:  int(p.ReactionsCount),
				LastCommentedAt: ct.GenDateTime(p.LastCommentedAt.AsTime()),
				CreatedAt:       ct.GenDateTime(p.CreatedAt.AsTime()),
				UpdatedAt:       ct.GenDateTime(p.UpdatedAt.AsTime()),
				LikedByUser:     p.LikedByUser,
				ImageId:         ct.Id(p.ImageId),
				ImageUrl:        p.ImageUrl,
			}
			postsResponse = append(postsResponse, post)
		}

		err = utils.WriteJSON(ctx, w, http.StatusOK, postsResponse)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, fmt.Sprintf("failed to send group %v posts: %v", groupId, err.Error()))
			return
		}

	}
}
