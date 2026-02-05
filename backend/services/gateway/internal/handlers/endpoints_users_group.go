package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"social-network/shared/gen-go/common"
	"social-network/shared/gen-go/media"
	"social-network/shared/gen-go/users"
	ct "social-network/shared/go/ct"
	utils "social-network/shared/go/http-utils"
	"social-network/shared/go/jwt"
	"social-network/shared/go/models"
	"time"
)

// ADD IMAGE UPLOAD
func (s *Handlers) createGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type createGroupData struct {
			GroupTitle       string `json:"group_title"`
			GroupDescription string `json:"group_description"`

			GroupImageName string `json:"group_image_name"`
			GroupImageSize int64  `json:"group_image_size"`
			GroupImageType string `json:"group_image_type"`
		}

		httpReq := createGroupData{}

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

		var GroupImageId ct.Id
		var uploadURL string

		if httpReq.GroupImageSize != 0 {
			exp := time.Duration(10 * time.Minute).Seconds()
			mediaRes, err := s.MediaService.UploadImage(r.Context(), &media.UploadImageRequest{
				Filename:          httpReq.GroupImageName,
				MimeType:          httpReq.GroupImageType,
				SizeBytes:         httpReq.GroupImageSize,
				Visibility:        media.FileVisibility_PUBLIC,
				Variants:          []media.FileVariant{media.FileVariant_SMALL},
				ExpirationSeconds: int64(exp),
			})
			if err != nil {
				utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
				return
			}
			GroupImageId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		createGroupRequest := users.CreateGroupRequest{
			OwnerId:          claims.UserId,
			GroupTitle:       httpReq.GroupTitle,
			GroupDescription: httpReq.GroupDescription,
			GroupImageId:     GroupImageId.Int64(),
		}

		groupId, err := s.UsersService.CreateGroup(ctx, &createGroupRequest)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not create group: "+err.Error())
			return
		}

		type createGroupDataResponse struct {
			GroupId   ct.Id `json:"group_id"`
			FileId    ct.Id
			UploadUrl string
		}

		resp := createGroupDataResponse{
			GroupId:   ct.Id(groupId.Value),
			FileId:    GroupImageId,
			UploadUrl: uploadURL,
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) updateGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type updateGroupData struct {
			GroupId          ct.Id
			GroupTitle       string `json:"group_title"`
			GroupDescription string `json:"group_description"`
			GroupImageId     ct.Id  `json:"group_image_id" validate:"nullable"`
			DeleteImage      bool   `json:"delete_image"`

			GroupImageName string `json:"group_image_name"`
			GroupImageSize int64  `json:"group_image_size"`
			GroupImageType string `json:"group_image_type"`
		}

		httpReq := updateGroupData{}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
		if err := decoder.Decode(&httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		var err error
		httpReq.GroupId, err = utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		if err := ct.ValidateStruct(httpReq); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, err.Error())
			return
		}

		var groupImageId ct.Id
		if httpReq.GroupImageId > 0 {
			groupImageId = httpReq.GroupImageId
		}
		var uploadURL string

		if httpReq.GroupImageSize != 0 {
			exp := time.Duration(10 * time.Minute).Seconds()
			mediaRes, err := s.MediaService.UploadImage(r.Context(), &media.UploadImageRequest{
				Filename:          httpReq.GroupImageName,
				MimeType:          httpReq.GroupImageType,
				SizeBytes:         httpReq.GroupImageSize,
				Visibility:        media.FileVisibility_PUBLIC,
				Variants:          []media.FileVariant{media.FileVariant_SMALL},
				ExpirationSeconds: int64(exp),
			})
			if err != nil {
				utils.ErrorJSON(ctx, w, http.StatusInternalServerError, err.Error())
				return
			}
			groupImageId = ct.Id(mediaRes.FileId)
			uploadURL = mediaRes.GetUploadUrl()
		}

		updateGroupRequest := users.UpdateGroupRequest{
			RequesterId:      claims.UserId,
			GroupId:          httpReq.GroupId.Int64(),
			GroupTitle:       httpReq.GroupTitle,
			GroupDescription: httpReq.GroupDescription,
			GroupImageId:     groupImageId.Int64(),
			DeleteImage:      httpReq.DeleteImage,
		}

		_, err = s.UsersService.UpdateGroup(ctx, &updateGroupRequest)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not update group: "+err.Error())
			return
		}

		type createGroupDataResponse struct {
			GroupId   ct.Id `json:"group_id"`
			FileId    ct.Id
			UploadUrl string
		}

		resp := createGroupDataResponse{
			GroupId:   ct.Id(httpReq.GroupId),
			FileId:    groupImageId,
			UploadUrl: uploadURL,
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) getAllGroupsPaginated() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		v := r.URL.Query()
		limit, err1 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err2 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := users.Pagination{
			UserId: claims.UserId,
			Limit:  limit,
			Offset: offset,
		}

		grpcResp, err := s.UsersService.GetAllGroupsPaginated(ctx, &req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch groups: "+err.Error())
			return
		}

		resp := make([]models.Group, 0, len(grpcResp.GroupArr))
		for _, group := range grpcResp.GroupArr {
			newGroup := models.Group{
				GroupId:          ct.Id(group.GroupId),
				GroupOwnerId:     ct.Id(group.GroupOwnerId),
				GroupTitle:       ct.Title(group.GroupTitle),
				GroupDescription: ct.About(group.GroupDescription),
				GroupImage:       ct.Id(group.GroupImageId),
				GroupImageURL:    group.GroupImageUrl,
				MembersCount:     group.MembersCount,
				IsMember:         group.IsMember,
				IsOwner:          group.IsOwner,
				PendingRequest:   group.PendingRequest,
				PendingInvite:    group.PendingInvite,
			}
			resp = append(resp, newGroup)
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) getGroupInfo() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		var err error
		groupId, err := utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.GeneralGroupRequest{
			GroupId: groupId.Int64(),
			UserId:  claims.UserId,
		}

		grpcResp, err := s.UsersService.GetGroupInfo(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch group info: "+err.Error())
			return
		}

		resp := models.Group{
			GroupId:          ct.Id(grpcResp.GroupId),
			GroupOwnerId:     ct.Id(grpcResp.GroupOwnerId),
			GroupTitle:       ct.Title(grpcResp.GroupTitle),
			GroupDescription: ct.About(grpcResp.GroupDescription),
			GroupImage:       ct.Id(grpcResp.GroupImageId),
			GroupImageURL:    grpcResp.GroupImageUrl,
			MembersCount:     grpcResp.MembersCount,
			IsMember:         grpcResp.IsMember,
			IsOwner:          grpcResp.IsOwner,
			PendingRequest:   grpcResp.PendingRequest,
			PendingInvite:    grpcResp.PendingInvite,
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) getGroupMembers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		req := &users.GroupMembersRequest{
			UserId:  claims.UserId,
			GroupId: groupId.Int64(),
			Limit:   limit,
			Offset:  offset,
		}

		grpcResp, err := s.UsersService.GetGroupMembers(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch group members: "+err.Error())
			return
		}

		resp := models.GroupUsers{}

		for _, group := range grpcResp.GroupUserArr {
			newGroup := models.GroupUser{
				UserId:    ct.Id(group.UserId),
				Username:  ct.Username(group.Username),
				AvatarId:  ct.Id(group.Avatar),
				AvatarUrl: group.AvatarUrl,
				GroupRole: group.GroupRole,
			}
			resp.GroupUsers = append(resp.GroupUsers, newGroup)
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) getUserGroupsPaginated() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		v := r.URL.Query()
		limit, err1 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err2 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.Pagination{
			UserId: claims.UserId,
			Limit:  limit,
			Offset: offset,
		}

		grpcResp, err := s.UsersService.GetUserGroupsPaginated(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch user groups: "+err.Error())
			return
		}

		resp := make([]models.Group, 0, len(grpcResp.GroupArr))
		for _, group := range grpcResp.GroupArr {
			newGroup := models.Group{
				GroupId:          ct.Id(group.GroupId),
				GroupOwnerId:     ct.Id(group.GroupOwnerId),
				GroupTitle:       ct.Title(group.GroupTitle),
				GroupDescription: ct.About(group.GroupDescription),
				GroupImage:       ct.Id(group.GroupImageId),
				GroupImageURL:    group.GroupImageUrl,
				MembersCount:     group.MembersCount,
				IsMember:         group.IsMember,
				IsOwner:          group.IsOwner,
				PendingRequest:   group.PendingRequest,
				PendingInvite:    group.PendingInvite,
			}
			resp = append(resp, newGroup)
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

// owner to accept or decline requests
func (s *Handlers) handleGroupJoinRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		body, err := utils.JSON2Struct(&models.HandleJoinRequest{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}

		body.GroupId, err = utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.HandleJoinRequest{
			OwnerId:     claims.UserId,
			GroupId:     body.GroupId.Int64(),
			RequesterId: body.RequesterId.Int64(),
			Accepted:    body.Accepted,
		}

		_, err = s.UsersService.HandleGroupJoinRequest(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not handle group join request: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

func (s *Handlers) cancelGroupJoinRequest() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		groupId, err := utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.GroupJoinRequest{
			GroupId:     groupId.Int64(),
			RequesterId: claims.UserId,
		}

		_, err = s.UsersService.CancelJoinGroupRequest(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not cancel group join request: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

func (s *Handlers) inviteToGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		body, err := utils.JSON2Struct(&models.InviteToGroupReq{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}

		body.GroupId, err = utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.InviteToGroupRequest{
			InviterId: claims.UserId,
			InvitedIds: &common.UserIds{
				Values: body.InvitedIds.Int64(),
			},
			GroupId: body.GroupId.Int64(),
		}

		_, err = s.UsersService.InviteToGroup(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not invite user to group: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

func (s *Handlers) leaveGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		// body, err := utils.JSON2Struct(&models.GroupJoinRequest{}, r)
		// if err != nil {
		// 	utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
		// 	return
		// }
		body := models.GeneralGroupReq{}

		var err error
		body.GroupId, err = utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.GeneralGroupRequest{
			UserId:  claims.UserId,
			GroupId: body.GroupId.Int64(),
		}

		_, err = s.UsersService.LeaveGroup(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not leave group: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

func (s *Handlers) removeFromGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		body, err := utils.JSON2Struct(&models.RemoveFromGroupRequest{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}

		body.GroupId, err = utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.RemoveFromGroupRequest{
			GroupId:  body.GroupId.Int64(),
			MemberId: body.MemberId.Int64(),
			OwnerId:  claims.UserId,
		}

		_, err = s.UsersService.RemoveFromGroup(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not leave group: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

// request to join a group
func (s *Handlers) requestJoinGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		// body, err := utils.JSON2Struct(&models.GroupJoinRequest{}, r)
		// if err != nil {
		// 	utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
		// 	return
		// }
		body := models.GroupJoinRequest{}

		var err error
		body.GroupId, err = utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.GroupJoinRequest{
			RequesterId: claims.UserId,
			GroupId:     body.GroupId.Int64(),
		}

		_, err = s.UsersService.RequestJoinGroup(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not process join request: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

// accept or decline an invitation to
func (s *Handlers) respondToGroupInvite() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		type reqBody struct {
			GroupId ct.Id `json:"group_id"`
			Accept  bool  `json:"accept"`
		}

		body, err := utils.JSON2Struct(&reqBody{}, r)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "Bad JSON data received")
			return
		}

		body.GroupId, err = utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.HandleGroupInviteRequest{
			InvitedId: claims.UserId,
			GroupId:   body.GroupId.Int64(),
			Accepted:  body.Accept,
		}

		_, err = s.UsersService.RespondToGroupInvite(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not respond to invite: "+err.Error())
			return
		}

		utils.WriteJSON(ctx, w, http.StatusOK, nil)
	}
}

func (s *Handlers) searchGroups() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		v := r.URL.Query()
		query, err1 := utils.ParamGet(v, "query", "", true)
		limit, err2 := utils.ParamGet(v, "limit", int32(1), false)
		offset, err3 := utils.ParamGet(v, "offset", int32(0), false)
		if err := errors.Join(err1, err2, err3); err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.GroupSearchRequest{
			SearchTerm: query,
			Limit:      limit,
			Offset:     offset,
			UserId:     claims.UserId,
		}

		grpcResp, err := s.UsersService.SearchGroups(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not search groups: "+err.Error())
			return
		}

		resp := models.Groups{
			Groups: make([]models.Group, 0, len(grpcResp.GroupArr)),
		}
		for _, group := range grpcResp.GroupArr {
			newGroup := models.Group{
				GroupId:          ct.Id(group.GroupId),
				GroupOwnerId:     ct.Id(group.GroupOwnerId),
				GroupTitle:       ct.Title(group.GroupTitle),
				GroupDescription: ct.About(group.GroupDescription),
				GroupImage:       ct.Id(group.GroupImageId),
				GroupImageURL:    group.GroupImageUrl,
				MembersCount:     group.MembersCount,
				IsMember:         group.IsMember,
				IsOwner:          group.IsOwner,
				PendingRequest:   group.PendingRequest,
				PendingInvite:    group.PendingInvite,
			}

			resp.Groups = append(resp.Groups, newGroup)
		}

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) getPendingGroupJoinRequests() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		req := &users.GroupMembersRequest{
			UserId:  claims.UserId,
			GroupId: groupId.Int64(),
			Limit:   limit,
			Offset:  offset,
		}

		grpcResp, err := s.UsersService.GetPendingGroupJoinRequests(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch group members: "+err.Error())
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

func (s *Handlers) getPendingGroupJoinRequestsCount() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		claims, ok := utils.GetValue[jwt.Claims](r, ct.ClaimsKey)
		if !ok {
			panic(1)
		}

		groupId, err := utils.PathValueGet(r, "group_id", ct.Id(0), true)
		if err != nil {
			utils.ErrorJSON(ctx, w, http.StatusBadRequest, "bad url params: "+err.Error())
			return
		}

		req := &users.GeneralGroupRequest{
			GroupId: groupId.Int64(),
			UserId:  claims.UserId,
		}

		grpcResp, err := s.UsersService.GetPendingGroupJoinRequestsCount(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch group members: "+err.Error())
			return
		}

		resp := grpcResp.Id

		utils.WriteJSON(ctx, w, http.StatusOK, resp)
	}
}

func (s *Handlers) GetFollowersNotInvitedToGroup() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
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

		req := &users.GroupMembersRequest{
			UserId:  claims.UserId,
			GroupId: groupId.Int64(),
			Limit:   limit,
			Offset:  offset,
		}

		grpcResp, err := s.UsersService.GetFollowersNotInvitedToGroup(ctx, req)
		if err != nil {
			utils.ReturnHttpError(ctx, w, err)
			//utils.ErrorJSON(ctx, w, http.StatusInternalServerError, "Could not fetch group members: "+err.Error())
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
