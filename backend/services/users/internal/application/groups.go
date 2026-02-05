package application

import (
	"context"
	"fmt"
	ds "social-network/services/users/internal/db/dbservice"
	"social-network/shared/gen-go/media"
	notifpb "social-network/shared/gen-go/notifications"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"github.com/jackc/pgx/v5/pgconn"
)

func (s *Application) GetAllGroupsPaginated(ctx context.Context, req models.Pagination) ([]models.Group, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.Group{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	//paginated (sorting by most members first)
	rows, err := s.db.GetAllGroups(ctx, ds.GetAllGroupsParams{
		Offset: req.Offset.Int32(),
		Limit:  req.Limit.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	groups := make([]models.Group, 0, len(rows))
	var imageIds ct.Ids

	for _, r := range rows {
		userInfo, err := s.userInRelationToGroup(ctx, models.GeneralGroupReq{
			GroupId: ct.Id(r.ID),
			UserId:  req.UserId,
		})
		if err != nil {
			return nil, err
		}

		groups = append(groups, models.Group{
			GroupId:          ct.Id(r.ID),
			GroupOwnerId:     ct.Id(r.GroupOwner),
			GroupTitle:       ct.Title(r.GroupTitle),
			GroupDescription: ct.About(r.GroupDescription),
			GroupImage:       ct.Id(r.GroupImageID),
			MembersCount:     r.MembersCount,
			IsMember:         userInfo.isMember,
			IsOwner:          userInfo.isOwner,
			PendingRequest:   userInfo.pendingRequest,
			PendingInvite:    userInfo.pendingInvite,
		})
		if r.GroupImageID > 0 {
			imageIds = append(imageIds, ct.Id(r.GroupImageID))
		}

	}

	//get image urls
	if len(imageIds) > 0 {
		imageMap, failedImageIds, err := s.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_SMALL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning
		} else {
			for i := range groups {
				groups[i].GroupImageURL = imageMap[groups[i].GroupImage.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return groups, nil
}

func (s *Application) GetUserGroupsPaginated(ctx context.Context, req models.Pagination) ([]models.Group, error) {
	input := fmt.Sprintf("%#v", req)

	//paginated (joined latest first)
	if err := ct.ValidateStruct(req); err != nil {
		return []models.Group{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	rows, err := s.db.GetUserGroups(ctx, ds.GetUserGroupsParams{
		GroupOwner: req.UserId.Int64(),
		Limit:      req.Limit.Int32(),
		Offset:     req.Offset.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	groups := make([]models.Group, 0, len(rows))
	var imageIds ct.Ids

	for _, r := range rows {
		pendingInfo, err := s.isGroupMembershipPending(ctx, models.GeneralGroupReq{
			GroupId: ct.Id(r.GroupID),
			UserId:  req.UserId,
		})
		if err != nil {
			return nil, ce.Wrap(nil, err)
		}
		groups = append(groups, models.Group{
			GroupId:          ct.Id(r.GroupID),
			GroupOwnerId:     ct.Id(r.GroupOwner),
			GroupTitle:       ct.Title(r.GroupTitle),
			GroupDescription: ct.About(r.GroupDescription),
			GroupImage:       ct.Id(r.GroupImageID),
			MembersCount:     r.MembersCount,
			IsMember:         r.IsMember,
			IsOwner:          r.IsOwner,
			PendingRequest:   pendingInfo.pendingRequest,
			PendingInvite:    pendingInfo.pendingInvite,
		})
		if r.GroupImageID > 0 {
			imageIds = append(imageIds, ct.Id(r.GroupImageID))
		}
	}

	//get image urls
	if len(imageIds) > 0 {
		imageMap, failedImageIds, err := s.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_SMALL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning
		} else {
			for i := range groups {
				groups[i].GroupImageURL = imageMap[groups[i].GroupImage.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return groups, nil
}

func (s *Application) GetGroupInfo(ctx context.Context, req models.GeneralGroupReq) (models.Group, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return models.Group{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	row, err := s.db.GetGroupInfo(ctx, req.GroupId.Int64())
	if err != nil {
		return models.Group{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	group := models.Group{
		GroupId:          ct.Id(row.ID),
		GroupOwnerId:     ct.Id(row.GroupOwner),
		GroupTitle:       ct.Title(row.GroupTitle),
		GroupDescription: ct.About(row.GroupDescription),
		GroupImage:       ct.Id(row.GroupImageID),
		MembersCount:     row.MembersCount,
	}
	userInfo, err := s.userInRelationToGroup(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.UserId,
	})
	if err != nil {
		return models.Group{}, ce.Wrap(nil, err)
	}
	group.IsMember = userInfo.isMember
	group.IsOwner = userInfo.isOwner
	group.PendingRequest = userInfo.pendingRequest
	group.PendingInvite = userInfo.pendingInvite

	if group.GroupImage > 0 {
		imageUrl, err := s.mediaRetriever.GetImage(ctx, group.GroupImage.Int64(), media.FileVariant_SMALL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", group.GroupImage, "error", err.Error()) //log error instead of returning
			s.removeFailedImage(ctx, err, group.GroupImage.Int64())
		} else {
			group.GroupImageURL = imageUrl
		}
	}

	return group, nil

}

func (s *Application) GetGroupMembers(ctx context.Context, req models.GroupMembersReq) ([]models.GroupUser, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.GroupUser{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	//check request comes from member
	isMember, err := s.IsGroupMember(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.UserId,
	})
	if err != nil {
		return nil, ce.Wrap(nil, err)
	}
	if !isMember {
		return nil, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not a member of group %v", req.UserId, req.GroupId), input).WithPublic("permission denied")
	}

	//paginated (newest first)
	rows, err := s.db.GetGroupMembers(ctx, ds.GetGroupMembersParams{
		GroupID: req.GroupId.Int64(),
		Limit:   req.Limit.Int32(),
		Offset:  req.Offset.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	members := make([]models.GroupUser, 0, len(rows))
	var imageIds ct.Ids

	for _, r := range rows {
		var role string
		if r.Role.Valid {
			role = string(r.Role.GroupRole)
		}

		members = append(members, models.GroupUser{
			UserId:    ct.Id(r.ID),
			Username:  ct.Username(r.Username),
			AvatarId:  ct.Id(r.AvatarID),
			GroupRole: role,
		})
		if r.AvatarID > 0 {
			imageIds = append(imageIds, ct.Id(r.AvatarID))
		}
	}

	//get avatar urls
	if len(imageIds) > 0 {
		avatarMap, failedImageIds, err := s.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_THUMBNAIL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning
		} else {
			for i := range members {
				members[i].AvatarUrl = avatarMap[members[i].AvatarId.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return members, nil
}

// intentionally doesn't return avatar urls to avoid redundant calls
func (s *Application) GetAllGroupMemberIds(ctx context.Context, req models.GroupId) (ct.Ids, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateBatch(ct.Id(req)); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	//paginated (newest first)
	rows, err := s.db.GetAllGroupMemberIds(ctx, ds.GetAllGroupMemberIdsParams{
		GroupID: ct.Id(req).Int64(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	members := make([]ct.Id, 0, len(rows))

	for _, r := range rows {
		members = append(members, ct.Id(r.ID))
	}

	tele.Debug(ctx, "Group @1 member ids: @2", "groupId", req, "memberIds", members)

	return members, nil
}

func (s *Application) SearchGroups(ctx context.Context, req models.GroupSearchReq) ([]models.Group, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.Group{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	//weighted (title more important than description)
	//paginated (most members first)
	rows, err := s.db.SearchGroups(ctx, ds.SearchGroupsParams{
		Query:  req.SearchTerm.String(),
		UserID: req.UserId.Int64(),
		Limit:  req.Limit.Int32(),
		Offset: req.Offset.Int32(),
	})
	if err != nil {
		return []models.Group{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if len(rows) == 0 {
		return []models.Group{}, nil
	}

	groups := make([]models.Group, 0, len(rows))
	var imageIds ct.Ids

	for _, r := range rows {
		pendingInfo, err := s.isGroupMembershipPending(ctx, models.GeneralGroupReq{
			GroupId: ct.Id(r.ID),
			UserId:  req.UserId,
		})
		if err != nil {
			return nil, ce.Wrap(nil, err)
		}
		groups = append(groups, models.Group{
			GroupId:          ct.Id(r.ID),
			GroupOwnerId:     ct.Id(r.GroupOwner),
			GroupTitle:       ct.Title(r.GroupTitle),
			GroupDescription: ct.About(r.GroupDescription),
			GroupImage:       ct.Id(r.GroupImageID),
			MembersCount:     r.MembersCount,
			IsMember:         r.IsMember,
			IsOwner:          r.IsOwner,
			PendingRequest:   pendingInfo.pendingRequest,
			PendingInvite:    pendingInfo.pendingInvite,
		})
		if r.GroupImageID > 0 {
			imageIds = append(imageIds, ct.Id(r.GroupImageID))
		}
	}

	//get image urls
	if len(imageIds) > 0 {
		imageMap, failedImageIds, err := s.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_SMALL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning
		} else {
			for i := range groups {
				groups[i].GroupImageURL = imageMap[groups[i].GroupImage.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return groups, nil

}

func (s *Application) InviteToGroup(ctx context.Context, req models.InviteToGroupReq) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	//check request comes from member
	isMember, err := s.IsGroupMember(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.InviterId,
	})
	if err != nil {
		return ce.Wrap(nil, err)
	}
	if !isMember {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not a member of group %v", req.InviterId, req.GroupId), input).WithPublic("permission denied")
	}

	err = s.db.SendGroupInvites(ctx, ds.SendGroupInvitesParams{
		GroupID:     req.GroupId.Int64(),
		SenderID:    req.InviterId.Int64(),
		ReceiverIDs: req.InvitedIds.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	// ============== create notification ===============================

	// create notification (waiting for batch notification)
	inviter, err := s.GetBasicUserInfo(ctx, req.InviterId)
	if err != nil {
		tele.Error(ctx, "Could not get basic user info for user @1 for invite to group notif: @2", "groupId", req.GroupId, "error", err.Error())
	}
	group, err := s.db.GetGroupBasicInfo(ctx, req.GroupId.Int64())
	if err != nil {
		tele.Error(ctx, "Could not get basic group info for group @1 for invite to group notif: @2", "groupId", req.GroupId, "error", err.Error())
	}
	event := &notifpb.NotificationEvent{
		EventType: notifpb.EventType_GROUP_INVITE_CREATED,
		Payload: &notifpb.NotificationEvent_GroupInviteCreated{
			GroupInviteCreated: &notifpb.GroupInviteCreated{
				InvitedUserId:   req.InvitedIds.Int64(),
				InviterUserId:   req.InviterId.Int64(),
				GroupId:         req.GroupId.Int64(),
				GroupName:       group.GroupTitle,
				InviterUsername: inviter.Username.String(),
			},
		},
	}

	if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
		tele.Error(ctx, "failed to send invite to group notification: @1", "error", err.Error())
	}
	tele.Info(ctx, "invite to group notification event created")

	return nil
}

// SKIP GRPC FOR NOW
// func (s *Application) CancelInviteToGroup(ctx context.Context, req models.InviteToGroupReq) error {
// 	if err := ct.ValidateStruct(req); err != nil {
// 		return err
// 	}
// 	err := s.db.CancelGroupInvite(ctx, ds.CancelGroupInviteParams{
// 		GroupID:    req.GroupId.Int64(),
// 		ReceiverID: req.InvitedId.Int64(),
// 		SenderID:   req.InviterId.Int64(),
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	//TODO REMOVE NOTIFICATION EVENT
// 	return nil
// }

func (s *Application) RequestJoinGroup(ctx context.Context, req models.GroupJoinRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	err := s.db.SendGroupJoinRequest(ctx, ds.SendGroupJoinRequestParams{
		GroupID: req.GroupId.Int64(),
		UserID:  req.RequesterId.Int64(),
	})

	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	//create notification
	requester, err := s.GetBasicUserInfo(ctx, req.RequesterId)
	if err != nil {
		tele.Error(ctx, "Could not get basic user info for id @1 for request join group notif: @2", "userId", req.RequesterId, "error", err.Error())
	}
	group, err := s.db.GetGroupBasicInfo(ctx, req.GroupId.Int64())
	if err != nil {
		tele.Error(ctx, "Could not get basic group info for id @1 for request join group notif: @2", "userId", req.GroupId, "error", err.Error())
	}

	event := &notifpb.NotificationEvent{
		EventType: notifpb.EventType_GROUP_JOIN_REQUEST_CREATED,
		Payload: &notifpb.NotificationEvent_GroupJoinRequestCreated{
			GroupJoinRequestCreated: &notifpb.GroupJoinRequestCreated{
				GroupOwnerId:      group.GroupOwner,
				RequesterUserId:   req.RequesterId.Int64(),
				GroupId:           req.GroupId.Int64(),
				GroupName:         group.GroupTitle,
				RequesterUsername: requester.Username.String(),
			},
		},
	}

	if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
		tele.Error(ctx, "failed to send request join group notification: @1", "error", err.Error())
	}
	tele.Info(ctx, "request join group notification event created")

	return nil
}

func (s *Application) CancelJoinGroupRequest(ctx context.Context, req models.GroupJoinRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	err := s.db.CancelGroupJoinRequest(ctx, ds.CancelGroupJoinRequestParams{
		GroupID: req.GroupId.Int64(),
		UserID:  req.RequesterId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	//remove notification
	group, err := s.db.GetGroupBasicInfo(ctx, req.GroupId.Int64())
	if err != nil {
		tele.Error(ctx, "Could not get basic group info for id @1 for cancel request join group notif: @2", "userId", req.GroupId, "error", err.Error())
	}

	event := &notifpb.NotificationEvent{
		EventType: notifpb.EventType_GROUP_JOIN_REQUEST_CANCELLED,
		Payload: &notifpb.NotificationEvent_GroupJoinRequestCancelled{
			GroupJoinRequestCancelled: &notifpb.GroupJoinRequestCancelled{
				GroupOwnerId:    group.GroupOwner,
				RequesterUserId: req.RequesterId.Int64(),
				GroupId:         req.GroupId.Int64(),
			},
		},
	}

	if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
		tele.Error(ctx, "failed to send cancel request to join group notification: @1", "error", err.Error())
	}
	tele.Info(ctx, "cancel request to join group notification event created")
	return nil
}

func (s *Application) RespondToGroupInvite(ctx context.Context, req models.HandleGroupInviteRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	inviterId, err := s.db.GetGroupInviterId(ctx, ds.GetGroupInviterIdParams{
		GroupID:    req.GroupId.Int64(),
		ReceiverID: req.InvitedId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if req.Accepted {

		err := s.db.AcceptGroupInvite(ctx, ds.AcceptGroupInviteParams{
			GroupID:    req.GroupId.Int64(),
			ReceiverID: req.InvitedId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)

		}
		//create notification
		invited, err := s.GetBasicUserInfo(ctx, req.InvitedId)
		if err != nil {
			tele.Error(ctx, "Could not get basic user info for id @1 for group invite accepted notif: @2", "userId", req.InvitedId, "error", err.Error())
		}
		group, err := s.db.GetGroupBasicInfo(ctx, req.GroupId.Int64())
		if err != nil {
			tele.Error(ctx, "Could not get basic group info for id @1 for group invite accepted notif: @2", "userId", req.GroupId, "error", err.Error())
		}

		event := &notifpb.NotificationEvent{
			EventType: notifpb.EventType_GROUP_INVITE_ACCEPTED,
			Payload: &notifpb.NotificationEvent_GroupInviteAccepted{
				GroupInviteAccepted: &notifpb.GroupInviteAccepted{
					InviterUserId:   inviterId,
					InvitedUserId:   req.InvitedId.Int64(),
					GroupId:         req.GroupId.Int64(),
					InvitedUsername: invited.Username.String(),
					GroupName:       group.GroupTitle,
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to send group invite accepted notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "group invite accepted notification event created")

	} else {
		err := s.db.DeclineGroupInvite(ctx, ds.DeclineGroupInviteParams{
			GroupID:    req.GroupId.Int64(),
			ReceiverID: req.InvitedId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}
		//create notification
		invited, err := s.GetBasicUserInfo(ctx, req.InvitedId)
		if err != nil {
			tele.Error(ctx, "Could not get basic user info for id @1 for group invite rejected notif: @2", "userId", req.InvitedId, "error", err.Error())
		}
		group, err := s.db.GetGroupBasicInfo(ctx, req.GroupId.Int64())
		if err != nil {
			tele.Error(ctx, "Could not get basic group info for id @1 for group invite rejected notif: @2", "userId", req.GroupId, "error", err.Error())
		}
		event := &notifpb.NotificationEvent{
			EventType: notifpb.EventType_GROUP_INVITE_REJECTED,
			Payload: &notifpb.NotificationEvent_GroupInviteRejected{
				GroupInviteRejected: &notifpb.GroupInviteRejected{
					InviterUserId:   inviterId,
					InvitedUserId:   req.InvitedId.Int64(),
					GroupId:         req.GroupId.Int64(),
					InvitedUsername: invited.Username.String(),
					GroupName:       group.GroupTitle,
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to send group invite rejected notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "group invite rejected notification event created")
	}
	return nil
}

func (s *Application) HandleGroupJoinRequest(ctx context.Context, req models.HandleJoinRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	isOwner, err := s.isGroupOwner(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.OwnerId,
	})
	if err != nil {
		return ce.Wrap(nil, err)
	}
	if !isOwner {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not the owner of group %v", req.OwnerId, req.GroupId), input).WithPublic("permission denied")
	}

	if req.Accepted {
		err = s.db.AcceptGroupJoinRequest(ctx, ds.AcceptGroupJoinRequestParams{
			GroupID: req.GroupId.Int64(),
			UserID:  req.RequesterId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		//create notification
		group, err := s.db.GetGroupBasicInfo(ctx, req.GroupId.Int64())
		if err != nil {
			tele.Error(ctx, "Could not get basic group info for id @1 for group join request accepted notif: @2", "userId", req.GroupId, "error", err.Error())
		}

		event := &notifpb.NotificationEvent{
			EventType: notifpb.EventType_GROUP_JOIN_REQUEST_ACCEPTED,
			Payload: &notifpb.NotificationEvent_GroupJoinRequestAccepted{
				GroupJoinRequestAccepted: &notifpb.GroupJoinRequestAccepted{
					RequesterUserId: req.RequesterId.Int64(),
					GroupOwnerId:    req.OwnerId.Int64(),
					GroupId:         req.GroupId.Int64(),
					GroupName:       group.GroupTitle,
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to send group join request accepted notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "group join request accepted notification event created")

	} else {
		err = s.db.RejectGroupJoinRequest(ctx, ds.RejectGroupJoinRequestParams{
			GroupID: req.GroupId.Int64(),
			UserID:  req.RequesterId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}
		//create notification
		group, err := s.db.GetGroupBasicInfo(ctx, req.GroupId.Int64())
		if err != nil {
			tele.Error(ctx, "Could not get basic group info for id @1 for group join request rejected notif: @2", "userId", req.GroupId, "error", err.Error())
		}
		event := &notifpb.NotificationEvent{
			EventType: notifpb.EventType_GROUP_JOIN_REQUEST_REJECTED,
			Payload: &notifpb.NotificationEvent_GroupJoinRequestRejected{
				GroupJoinRequestRejected: &notifpb.GroupJoinRequestRejected{
					RequesterUserId: req.RequesterId.Int64(),
					GroupOwnerId:    req.GroupId.Int64(),
					GroupId:         req.GroupId.Int64(),
					GroupName:       group.GroupTitle,
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to send group join request accepted notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "group join request accepted notification event created")
	}

	return nil
}

func (s *Application) LeaveGroup(ctx context.Context, req models.GeneralGroupReq) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	//check request comes from member
	isMember, err := s.IsGroupMember(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.UserId,
	})
	if err != nil {
		return ce.Wrap(nil, err)
	}
	if !isMember {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not a member of group %v", req.UserId, req.GroupId), input).WithPublic("permission denied")
	}

	err = s.db.LeaveGroup(ctx, ds.LeaveGroupParams{
		GroupID: req.GroupId.Int64(),
		UserID:  req.UserId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	return nil
}

func (s *Application) RemoveFromGroup(ctx context.Context, req models.RemoveFromGroupRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	//check owner has indeed the owner role
	isOwner, err := s.isGroupOwner(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.OwnerId,
	})
	if err != nil {
		return ce.Wrap(nil, err)
	}
	if !isOwner {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not the owner of group %v", req.OwnerId, req.GroupId), input).WithPublic("permission denied")
	}

	err = s.LeaveGroup(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.MemberId,
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	return nil
}

func (s *Application) CreateGroup(ctx context.Context, req *models.CreateGroupRequest) (models.GroupId, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return 0, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	groupId, err := s.db.CreateGroup(ctx, ds.CreateGroupParams{
		GroupOwner:       req.OwnerId.Int64(),
		GroupTitle:       req.GroupTitle.String(),
		GroupDescription: req.GroupDescription.String(),
		GroupImageID:     req.GroupImage.Int64(),
	})
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" { // unique_violation
				return 0, ce.New(ce.ErrAlreadyExists, err, input).WithPublic("group already exists")
			}
		}
		return 0, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	return models.GroupId(groupId), nil
}

func (s *Application) UpdateGroup(ctx context.Context, req *models.UpdateGroupRequest) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	//check requester is owner
	isOwner, err := s.isGroupOwner(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.RequesterId,
	})
	if err != nil {
		return ce.Wrap(nil, err)
	}
	if !isOwner {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not the owner of group %v", req.RequesterId, req.GroupId), input).WithPublic("permission denied")
	}

	groupImageId := req.GroupImage.Int64()
	if req.DeleteImage {
		groupImageId = 0
	}

	rowsAffected, err := s.db.UpdateGroup(ctx, ds.UpdateGroupParams{
		ID:               req.GroupId.Int64(),
		GroupTitle:       req.GroupTitle.String(),
		GroupDescription: req.GroupDescription.String(),
		GroupImageID:     groupImageId,
	})

	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if rowsAffected != 1 {
		return ce.New(ce.ErrNotFound, fmt.Errorf("group %v was not found or has been deleted", req.GroupId), input).WithPublic("not found")
	}

	return nil

}

// NOT GRPC
func (s *Application) userInRelationToGroup(ctx context.Context, req models.GeneralGroupReq) (resp userInRelationToGroup, err error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return userInRelationToGroup{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	resp.isOwner, err = s.isGroupOwner(ctx, req)
	if err != nil {
		return userInRelationToGroup{}, ce.Wrap(nil, err)
	}
	resp.isMember, err = s.IsGroupMember(ctx, req)
	if err != nil {
		return userInRelationToGroup{}, ce.Wrap(nil, err)
	}
	pendingInfo, err := s.isGroupMembershipPending(ctx, req)
	if err != nil {
		return userInRelationToGroup{}, ce.Wrap(nil, err)
	}

	resp.pendingRequest = pendingInfo.pendingRequest
	resp.pendingInvite = pendingInfo.pendingInvite

	return resp, nil
}

// NOT GRPC
func (s *Application) isGroupOwner(ctx context.Context, req models.GeneralGroupReq) (bool, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return false, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	isOwner, err := s.db.IsUserGroupOwner(ctx, ds.IsUserGroupOwnerParams{
		ID:         req.GroupId.Int64(),
		GroupOwner: req.UserId.Int64(),
	})
	if err != nil {
		return false, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if !isOwner {
		return false, nil
	}
	return true, nil
}

func (s *Application) IsGroupMember(ctx context.Context, req models.GeneralGroupReq) (bool, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return false, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	isMember, err := s.db.IsUserGroupMember(ctx, ds.IsUserGroupMemberParams{
		GroupID: req.GroupId.Int64(),
		UserID:  req.UserId.Int64(),
	})
	if err != nil {
		return false, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if !isMember {
		return false, nil
	}
	return true, nil
}

// NOT GRPC
func (s *Application) isGroupMembershipPending(ctx context.Context, req models.GeneralGroupReq) (isMembershipPending, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return isMembershipPending{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	row, err := s.db.IsGroupMembershipPending(ctx, ds.IsGroupMembershipPendingParams{
		GroupID: req.GroupId.Int64(),
		UserID:  req.UserId.Int64(),
	})
	if err != nil {
		return isMembershipPending{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	return isMembershipPending{
		pendingRequest: row.HasPendingJoinRequest,
		pendingInvite:  row.HasPendingInvite,
	}, nil
}

func (s *Application) GetFollowersNotInvitedToGroup(ctx context.Context, req models.GroupMembersReq) ([]models.User, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	//check request comes from member
	isMember, err := s.IsGroupMember(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.UserId,
	})
	if err != nil {
		return nil, ce.Wrap(nil, err)
	}
	if !isMember {
		return nil, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not a member of group %v", req.UserId, req.GroupId), input).WithPublic("permission denied")
	}

	//paginated, sorted by newest first
	rows, err := s.db.GetFollowersNotInvitedToGroup(ctx, ds.GetFollowersNotInvitedToGroupParams{
		UserId:  req.UserId.Int64(),
		GroupId: req.GroupId.Int64(),
		Limit:   int(req.Limit),
		Offset:  int(req.Offset),
	})
	if err != nil {
		return []models.User{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	users := make([]models.User, 0, len(rows))
	var imageIds ct.Ids
	for _, r := range rows {
		users = append(users, models.User{
			UserId:   ct.Id(r.Id),
			Username: ct.Username(r.Username),
			AvatarId: ct.Id(r.AvatarId),
		})
		if r.AvatarId > 0 {
			imageIds = append(imageIds, ct.Id(r.AvatarId))
		}
	}
	//get avatar urls
	if len(imageIds) > 0 {
		avatarMap, failedImageIds, err := s.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_THUMBNAIL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning
		} else {
			for i := range users {
				users[i].AvatarURL = avatarMap[users[i].AvatarId.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return users, nil

}

func (s *Application) GetPendingGroupJoinRequests(ctx context.Context, req models.GroupMembersReq) ([]models.User, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	isOwner, err := s.isGroupOwner(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.UserId,
	})
	if err != nil {
		return []models.User{}, ce.Wrap(nil, err)
	}
	if !isOwner {
		return []models.User{}, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not the owner of group %v", req.UserId, req.GroupId), input).WithPublic("permission denied")
	}

	//paginated, sorted by newest first
	rows, err := s.db.GetPendingGroupJoinRequests(ctx, ds.GetPendingGroupJoinRequestsParams{
		GroupId: req.GroupId.Int64(),
		Limit:   int(req.Limit),
		Offset:  int(req.Offset),
	})
	if err != nil {
		return []models.User{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	users := make([]models.User, 0, len(rows))
	var imageIds ct.Ids
	for _, r := range rows {
		users = append(users, models.User{
			UserId:   ct.Id(r.Id),
			Username: ct.Username(r.Username),
			AvatarId: ct.Id(r.AvatarId),
		})
		if r.AvatarId > 0 {
			imageIds = append(imageIds, ct.Id(r.AvatarId))
		}
	}
	//get avatar urls
	if len(imageIds) > 0 {
		avatarMap, failedImageIds, err := s.mediaRetriever.GetImages(ctx, imageIds, media.FileVariant_THUMBNAIL)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", imageIds, "error", err.Error()) //log error instead of returning
		} else {
			for i := range users {
				users[i].AvatarURL = avatarMap[users[i].AvatarId.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return users, nil

}

func (s *Application) GetPendingGroupJoinRequestsCount(ctx context.Context, req models.GroupJoinRequest) (int64, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return 0, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	isOwner, err := s.isGroupOwner(ctx, models.GeneralGroupReq{
		GroupId: req.GroupId,
		UserId:  req.RequesterId,
	})
	if err != nil {
		return 0, ce.Wrap(nil, err)
	}
	if !isOwner {
		return 0, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user %v is not the owner of group %v", req.RequesterId, req.GroupId), input).WithPublic("permission denied")
	}

	//paginated, sorted by newest first
	count, err := s.db.GetPendingGroupJoinRequestsCount(ctx, ds.GetPendingGroupJoinRequestsCountParams{
		GroupId: req.GroupId.Int64(),
	})
	if err != nil {
		return 0, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	return count, nil
}

func (s *Application) GetGroupBasicInfo(ctx context.Context, req models.GroupId) (models.Group, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateBatch(ct.Id(req)); err != nil {
		return models.Group{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	row, err := s.db.GetGroupBasicInfo(ctx, int64(req))
	if err != nil {
		return models.Group{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	group := models.Group{
		GroupId:          ct.Id(row.ID),
		GroupOwnerId:     ct.Id(row.GroupOwner),
		GroupTitle:       ct.Title(row.GroupTitle),
		GroupDescription: ct.About(row.GroupDescription),
		GroupImage:       ct.Id(row.GroupImageID),
	}
	return group, nil
}

// ---------------------------------------------------------------------
// low priority
// ---------------------------------------------------------------------
func DeleteGroup() {}

//called with group_id, owner_id
//returns success or error
//request needs to come from owner
//---------------------------------------------------------------------

//initiated by ownder
//SoftDeleteGroup

func TranferGroupOwnerShip() {}

//called with group_id,previous_owner_id, new_owner_id
//returns success or error
//request needs to come from previous owner (or admin - not implemented)
//---------------------------------------------------------------------
