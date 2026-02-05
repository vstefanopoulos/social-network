package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/users"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/gorpc"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
	"time"
)

func (h *Handlers) getUserProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "getUserProfile handler called")

		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}
		requesterId := int64(claims.UserId)

		userId, err := utils.PathValueGet(r, "user_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		grpcReq := users.GetUserProfileRequest{
			UserId:      userId.Int64(),
			RequesterId: requesterId,
		}

		grpcResp, err := h.UsersService.GetUserProfile(r.Context(), &grpcReq)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to get user info: "+err.Error())
			return
		}

		tele.Info(ctx, "retrieved user profile. @1", "grpcResp", grpcResp)

		type userProfile struct {
			UserId                        ct.Id          `json:"user_id"`
			Username                      ct.Username    `json:"username"`
			FirstName                     ct.Name        `json:"first_name"`
			LastName                      ct.Name        `json:"last_name"`
			DateOfBirth                   ct.DateOfBirth `json:"date_of_birth"`
			Avatar                        ct.Id          `json:"avatar,omitempty"`
			AvatarURL                     string         `json:"avatar_url,omitempty"`
			About                         ct.About       `json:"about,omitempty"`
			Public                        bool           `json:"public"`
			CreatedAt                     time.Time      `json:"created_at"`
			Email                         fmt.Stringer   `json:"email"`
			FollowersCount                int64          `json:"followers_count"`
			FollowingCount                int64          `json:"following_count"`
			GroupsCount                   int64          `json:"groups_count"`
			OwnedGroupsCount              int64          `json:"owned_groups_count"`
			ViewerIsFollowing             bool           `json:"viewer_is_following"`
			OwnProfile                    bool           `json:"own_profile"`
			IsPending                     bool           `json:"is_pending"`
			FollowRequestFromProfileOwner bool           `json:"follow_request_from_profile_owner"`
		}

		userProfileResponse := userProfile{
			UserId:                        ct.Id(grpcResp.UserId),
			Username:                      ct.Username(grpcResp.Username),
			FirstName:                     ct.Name(grpcResp.FirstName),
			LastName:                      ct.Name(grpcResp.LastName),
			DateOfBirth:                   ct.DateOfBirth(grpcResp.DateOfBirth.AsTime()),
			Avatar:                        ct.Id(grpcResp.Avatar),
			AvatarURL:                     grpcResp.AvatarUrl,
			About:                         ct.About(grpcResp.About),
			Public:                        grpcResp.Public,
			CreatedAt:                     grpcResp.CreatedAt.AsTime(),
			Email:                         ct.Email(grpcResp.Email),
			FollowersCount:                grpcResp.FollowersCount,
			FollowingCount:                grpcResp.FollowingCount,
			GroupsCount:                   grpcResp.GroupsCount,
			OwnedGroupsCount:              grpcResp.OwnedGroupsCount,
			ViewerIsFollowing:             grpcResp.ViewerIsFollowing,
			OwnProfile:                    grpcResp.OwnProfile,
			IsPending:                     grpcResp.IsPending,
			FollowRequestFromProfileOwner: grpcResp.FollowRequestFromProfileOwner,
		}

		tele.Info(ctx, "transformed profile struct. @1", "response", userProfileResponse)

		err = utils.WriteJSON(ctx, w, http.StatusOK, userProfileResponse)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "failed to send user info")
			return
		}

	}
}

func (s *Handlers) searchUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		v := r.URL.Query()
		query, err1 := utils.ParamGet(v, "query", "", true)
		limit, err2 := utils.ParamGet(v, "limit", int32(1), false)
		if err := errors.Join(err1, err2); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.UserSearchRequest{
			SearchTerm: query,
			Limit:      limit,
		}

		grpcResp, err := s.UsersService.SearchUsers(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not search users: "+err.Error())
			return
		}

		resp := models.Users{
			Users: make([]models.User, 0, len(grpcResp.Users)),
		}

		for _, user := range grpcResp.Users {
			newUser := models.User{
				UserId:    ct.Id(user.UserId),
				Username:  ct.Username(user.Username),
				AvatarId:  ct.Id(user.Avatar),
				AvatarURL: user.AvatarUrl,
			}

			resp.Users = append(resp.Users, newUser)
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) updateProfilePrivacy() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type reqBody struct {
			Public bool `json:"public"`
		}

		body, err := utils.JSON2Struct(&reqBody{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}

		req := &users.UpdateProfilePrivacyRequest{
			UserId: claims.UserId,
			Public: body.Public,
		}

		_, err = s.UsersService.UpdateProfilePrivacy(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not update privacy: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

func (s *Handlers) updateUserProfile() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type UpdateProfileJSONRequest struct {
			Username    ct.Username    `json:"username"`
			FirstName   ct.Name        `json:"first_name"`
			LastName    ct.Name        `json:"last_name"`
			DateOfBirth ct.DateOfBirth `json:"date_of_birth"`
			About       ct.About       `json:"about" validate:"nullable"`
			AvatarId    ct.Id          `json:"avatar_id" validate:"nullable"`
			DeleteImage bool           `json:"delete_image"`

			AvatarName string `json:"avatar_name"`
			AvatarSize int64  `json:"avatar_size"`
			AvatarType string `json:"avatar_type"`
		}

		httpReq := UpdateProfileJSONRequest{}

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

		var avatarId ct.Id
		if httpReq.AvatarId > 0 {
			avatarId = httpReq.AvatarId
		}
		var uploadURL string
		if httpReq.AvatarSize != 0 {
			exp := time.Duration(10 * time.Minute).Seconds()
			mediaRes, err := s.MediaService.UploadImage(r.Context(), &media.UploadImageRequest{
				Filename:          httpReq.AvatarName,
				MimeType:          httpReq.AvatarType,
				SizeBytes:         httpReq.AvatarSize,
				Visibility:        media.FileVisibility_PUBLIC,
				Variants:          []media.FileVariant{media.FileVariant_THUMBNAIL},
				ExpirationSeconds: int64(exp),
			})
			if err != nil {
				utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
				return
			}
			avatarId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		//MAKE GRPC REQUEST
		grpcRequest := &users.UpdateProfileRequest{
			UserId:      claims.UserId,
			Username:    httpReq.Username.String(),
			FirstName:   httpReq.FirstName.String(),
			LastName:    httpReq.LastName.String(),
			DateOfBirth: httpReq.DateOfBirth.ToProto(),
			Avatar:      avatarId.Int64(),
			About:       httpReq.About.String(),
			DeleteImage: httpReq.DeleteImage,
		}

		grpcResp, err := s.UsersService.UpdateUserProfile(ctx, grpcRequest)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not update profile: "+err.Error())
			return
		}

		type httpResponse struct {
			UserId    ct.Id
			FileId    ct.Id
			UploadUrl string
		}
		httpResp := httpResponse{
			UserId:    ct.Id(grpcResp.UserId),
			FileId:    avatarId,
			UploadUrl: uploadURL}

		utils.WriteJSON(ctx, w, http.StatusOK, httpResp)
	}
}

// Get endpoint to fetch username, avatarid and avatar thumbnail variant download url by user name.
// endpoint: /users/{user_id}/retrieve
func (h *Handlers) retrieveUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "retriveUser handler called")

		userId, err := utils.PathValueGet(r, "user_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}
		user, err := h.RetriveUsers.GetUser(ctx, userId)
		if err != nil {
			httpCode, _ := gorpc.Classify(err)
			err = ce.DecodeProto(err)
			tele.Error(ctx, "get BasicUserInfo error @1 @2", "request", userId, "error", err)
			utils.ErrorJSON(ctx, w, httpCode, err.Error())
			return
		}

		if err := utils.WriteJSON(ctx, w, http.StatusOK, user); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}

// Get endpoint to fetch username, avatarid and avatar thumbnail variant download url by user name.
// endpoint: /users/retrieve?id=1&id=2&id=3
func (h *Handlers) retrieveUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		tele.Info(ctx, "retriveUsers handler called")

		idsStrSlice := r.URL.Query()["id"]
		ids := make(ct.Ids, 0, len(idsStrSlice))

		for _, s := range idsStrSlice {
			id, err := ct.DecodeId(s)
			if err != nil {
				tele.Info(ctx, "invalid query param @1", "param", s)
			}
			ids = append(ids, ct.Id(id))
		}

		if len(ids) == 0 {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: ")
			return
		}

		user, err := h.RetriveUsers.GetUsers(ctx, ids)
		if err != nil {
			httpCode, _ := gorpc.Classify(err)
			err = ce.DecodeProto(err)
			tele.Error(ctx, "retrive Users error @1 @2", "request", ids, "error", err)
			utils.ErrorJSON(ctx, w, httpCode, err.Error())
			return
		}

		if err := utils.WriteJSON(ctx, w, http.StatusOK, user); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
		}
	}
}
