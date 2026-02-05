package handlers

import (
	"errors"
	"net/http"
	"social-network/shared/gen-go/common"
	"social-network/shared/gen-go/posts"
	"social-network/shared/gen-go/users"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func (s *Handlers) getFollowSuggestions() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}
		requesterId := int64(claims.UserId)

		req := wrapperspb.Int64Value{Value: requesterId}

		part1, err := s.UsersService.GetFollowSuggestions(ctx, &req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch suggestions from users: "+err.Error())
			return
		}

		part2, err := s.PostsService.SuggestUsersByPostActivity(ctx, &posts.SimpleIdReq{Id: requesterId})
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch suggestions from posts: "+err.Error())
			return
		}

		tele.Debug(ctx, "from users @1 @2", "part1", part1, "part2", part2)

		myMap := make(map[int64]*common.User)
		for _, user := range part1.Users {
			myMap[user.UserId] = user
		}
		for _, user := range part2.Users {
			myMap[user.UserId] = user
		}
		dedupedUsers := make([]models.User, 0, len(part1.Users)+len(part2.Users))
		for _, user := range myMap {
			newUser := models.User{
				UserId:    ct.Id(user.UserId),
				Username:  ct.Username(user.Username),
				AvatarId:  ct.Id(user.Avatar),
				AvatarURL: user.AvatarUrl,
			}
			dedupedUsers = append(dedupedUsers, newUser)
		}
		resp := models.Users{
			Users: dedupedUsers,
		}
		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) getFollowersPaginated() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		v := r.URL.Query()
		userId, err1 := utils.PathValueGet(r, "user_id", ct.Id(0), true)
		limit, err2 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err3 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2, err3); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := users.Pagination{
			UserId: userId.Int64(),
			Limit:  limit,
			Offset: offset,
		}

		grpcResp, err := s.UsersService.GetFollowersPaginated(ctx, &req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch followers: "+err.Error())
			return
		}

		resp := make([]models.User, 0, len(grpcResp.Users))
		for _, user := range grpcResp.Users {
			newUser := models.User{
				UserId:    ct.Id(user.UserId),
				Username:  ct.Username(user.Username),
				AvatarId:  ct.Id(user.Avatar),
				AvatarURL: user.AvatarUrl,
			}
			resp = append(resp, newUser)
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) getFollowingPaginated() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		// if !ok {
		// 	panic(1)
		// }

		v := r.URL.Query()
		userId, err1 := utils.PathValueGet(r, "user_id", ct.Id(0), true)
		limit, err2 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err3 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2, err3); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.Pagination{
			UserId: userId.Int64(),
			Limit:  limit,
			Offset: offset,
		}

		grpcResp, err := s.UsersService.GetFollowingPaginated(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch following users: "+err.Error())
			return
		}

		resp := &models.Users{}

		for _, grpcUser := range grpcResp.Users {
			user := models.User{
				UserId:    ct.Id(grpcUser.UserId),
				Username:  ct.Username(grpcUser.Username),
				AvatarId:  ct.Id(grpcUser.Avatar),
				AvatarURL: grpcUser.AvatarUrl,
			}
			resp.Users = append(resp.Users, user)
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) followUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		targetUserId, err := utils.PathValueGet(r, "user_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := users.FollowUserRequest{
			FollowerId:   claims.UserId,
			TargetUserId: targetUserId.Int64(),
		}

		resp, err := s.UsersService.FollowUser(ctx, &req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not follow user: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp) //TODO check if returned values need to be removed
	}
}

func (s *Handlers) handleFollowRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type reqBody struct {
			RequesterId ct.Id
			Accept      bool `json:"accept"`
		}

		body, err := utils.JSON2Struct(&reqBody{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}

		body.RequesterId, err = utils.PathValueGet(r, "user_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.HandleFollowRequestRequest{
			UserId:      claims.UserId,
			RequesterId: body.RequesterId.Int64(),
			Accept:      body.Accept,
		}

		_, err = s.UsersService.HandleFollowRequest(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not handle follow request: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

func (s *Handlers) unFollowUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		targetUserId, err := utils.PathValueGet(r, "user_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.FollowUserRequest{
			FollowerId:   claims.UserId,
			TargetUserId: targetUserId.Int64(),
		}

		_, err = s.UsersService.UnFollowUser(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not unfollow user: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}
