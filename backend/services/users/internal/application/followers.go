package application

import (
	"context"
	"errors"
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

func (s *Application) GetFollowersPaginated(ctx context.Context, req models.Pagination) ([]models.User, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}
	//paginated, sorted by newest first
	rows, err := s.db.GetFollowers(ctx, ds.GetFollowersParams{
		FollowingID: req.UserId.Int64(),
		Limit:       req.Limit.Int32(),
		Offset:      req.Offset.Int32(),
	})
	if err != nil {
		return []models.User{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	users := make([]models.User, 0, len(rows))
	var imageIds ct.Ids
	for _, r := range rows {
		users = append(users, models.User{
			UserId:   ct.Id(r.ID),
			Username: ct.Username(r.Username),
			AvatarId: ct.Id(r.AvatarID),
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
			for i := range users {
				users[i].AvatarURL = avatarMap[users[i].AvatarId.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return users, nil

}

func (s *Application) GetFollowingPaginated(ctx context.Context, req models.Pagination) ([]models.User, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return []models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	//paginated, sorted by newest first
	rows, err := s.db.GetFollowing(ctx, ds.GetFollowingParams{
		FollowerID: req.UserId.Int64(),
		Limit:      req.Limit.Int32(),
		Offset:     req.Offset.Int32(),
	})
	if err != nil {
		return []models.User{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	users := make([]models.User, 0, len(rows))
	var imageIds ct.Ids
	for _, r := range rows {
		users = append(users, models.User{
			UserId:   ct.Id(r.ID),
			Username: ct.Username(r.Username),
			AvatarId: ct.Id(r.AvatarID),
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
			for i := range users {
				users[i].AvatarURL = avatarMap[users[i].AvatarId.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return users, nil

}

func (s *Application) FollowUser(ctx context.Context, req models.FollowUserReq) (resp models.FollowUserResp, err error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return models.FollowUserResp{}, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}
	//status can be already_following, refollowed, followed, request_already_pending, request_resent,requested
	status, err := s.db.FollowUser(ctx, ds.FollowUserParams{
		PFollower: req.FollowerId.Int64(),
		PTarget:   req.TargetUserId.Int64(),
	})

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "22023": // invalid_parameter_value
				return models.FollowUserResp{}, ce.New(ce.ErrInvalidArgument, err, input).WithPublic("user cannot follow self")
			case "P0002": // custom "not found"
				return models.FollowUserResp{}, ce.New(ce.ErrNotFound, err, input).WithPublic("user not found")
			}
		}
		return models.FollowUserResp{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if status == "requested" || status == "request_resent" || status == "request_already_pending" { //Profile was private, request sent
		resp.IsPending = true
		resp.ViewerIsFollowing = false

		if status != "request_already_pending" {
			// =======================  create notification ====================================
			follower, err := s.GetBasicUserInfo(ctx, req.FollowerId)
			if err != nil {
				tele.Error(ctx, "Could not get basic user info for id @1 for follow request notif: @2", "userId", req.FollowerId, "error", err.Error())
			}

			// build the notification event
			event := &notifpb.NotificationEvent{
				EventType: notifpb.EventType_FOLLOW_REQUEST_CREATED,
				Payload: &notifpb.NotificationEvent_FollowRequestCreated{
					FollowRequestCreated: &notifpb.FollowRequestCreated{
						TargetUserId:      req.TargetUserId.Int64(),
						RequesterUserId:   req.FollowerId.Int64(),
						RequesterUsername: follower.Username.String(),
					},
				},
			}

			if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
				tele.Error(ctx, "failed to send new follow request notification: @1", "error", err.Error())
			}
			tele.Info(ctx, "Follow request notification event created")
		}
	} else {
		resp.IsPending = false
		resp.ViewerIsFollowing = true

		// =======================  create notification if not refollow ====================================

		if status == "followed" { //only first follow creates notification
			follower, err := s.GetBasicUserInfo(ctx, req.FollowerId)
			if err != nil {
				tele.Error(ctx, "Could not get basic user info for id @1 for follow request notif: @2", "userId", req.FollowerId, "error", err.Error())
			}

			event := &notifpb.NotificationEvent{
				EventType: notifpb.EventType_NEW_FOLLOWER_CREATED,
				Payload: &notifpb.NotificationEvent_NewFollowerCreated{
					NewFollowerCreated: &notifpb.NewFollowerCreated{
						TargetUserId:     req.TargetUserId.Int64(),
						FollowerUserId:   req.FollowerId.Int64(),
						FollowerUsername: follower.Username.String(),
					},
				},
			}

			if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
				tele.Error(ctx, "failed to send new follower notification: @1", "error", err.Error())
			}
			tele.Info(ctx, "new follower notification event created")
		}
	}

	return resp, nil
}

func (s *Application) UnFollowUser(ctx context.Context, req models.FollowUserReq) (err error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}
	//if already following, unfollows
	// if request pending, cancels request

	action, err := s.db.UnfollowUser(ctx, ds.UnfollowUserParams{
		FollowerID:  req.FollowerId.Int64(),
		FollowingID: req.TargetUserId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	switch action {
	case "unfollow":
		tele.Debug(ctx, "user @1 successfully unfollowed target user @2", "userId", req.FollowerId, "targetUserId", req.TargetUserId)

	case "cancel_request":
		tele.Debug(ctx, "user @1 successfully canceled follow request to target user @2", "userId", req.FollowerId, "targetUserId", req.TargetUserId)
		// cancel follow notification

		event := &notifpb.NotificationEvent{
			EventType: *notifpb.EventType_FOLLOW_REQUEST_CANCELLED.Enum(),
			Payload: &notifpb.NotificationEvent_FollowRequestCancelled{
				FollowRequestCancelled: &notifpb.FollowRequestCancelled{
					RequesterUserId: req.FollowerId.Int64(),
					TargetUserId:    req.TargetUserId.Int64(),
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to cancel follow request notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "cancel follow request accepted notification event created")

	default:
		//neither unfollow nor cancel_request was returned - no rows were affected
		// Log aggressively, but don't fail the user
		tele.Warn(ctx, "no rows affected")

	}

	return nil
}

func (s *Application) HandleFollowRequest(ctx context.Context, req models.HandleFollowRequestReq) error {
	input := fmt.Sprintf("%#v", req)

	var err error
	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	if req.Accept {
		err = s.db.AcceptFollowRequest(ctx, ds.AcceptFollowRequestParams{
			RequesterID: req.RequesterId.Int64(),
			TargetID:    req.UserId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		//create notification
		targetUser, err := s.GetBasicUserInfo(ctx, req.UserId)
		if err != nil {
			tele.Error(ctx, "Could not get basic user info for id @1 for follow request accepted notif: @2", "userId", req.UserId, "error", err.Error())
		}

		event := &notifpb.NotificationEvent{
			EventType: *notifpb.EventType_FOLLOW_REQUEST_ACCEPTED.Enum(),
			Payload: &notifpb.NotificationEvent_FollowRequestAccepted{
				FollowRequestAccepted: &notifpb.FollowRequestAccepted{
					RequesterUserId: req.RequesterId.Int64(),
					TargetUserId:    targetUser.UserId.Int64(),
					TargetUsername:  targetUser.Username.String(),
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to follow request accepted notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "follow request accepted notification event created")

	} else {
		err = s.db.RejectFollowRequest(ctx, ds.RejectFollowRequestParams{
			RequesterID: req.RequesterId.Int64(),
			TargetID:    req.UserId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}
		//create notification
		targetUser, err := s.GetBasicUserInfo(ctx, req.UserId)
		if err != nil {
			tele.Error(ctx, "Could not get basic user info for id @1 for follow request rejected notif: @2", "userId", req.UserId, "error", err.Error())
		}

		event := &notifpb.NotificationEvent{
			EventType: notifpb.EventType_FOLLOW_REQUEST_REJECTED,
			Payload: &notifpb.NotificationEvent_FollowRequestRejected{
				FollowRequestRejected: &notifpb.FollowRequestRejected{
					RequesterUserId: req.RequesterId.Int64(),
					TargetUserId:    targetUser.UserId.Int64(),
					TargetUsername:  targetUser.Username.String(),
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to send follow request rejected notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "follow request rejected notification event created")

	}
	return nil
}

// returns ids of people a user follows for posts service so that the feed can be fetched
func (s *Application) GetFollowingIds(ctx context.Context, userId ct.Id) ([]int64, error) {
	input := fmt.Sprintf("%#v", userId)

	if err := userId.Validate(); err != nil {
		return []int64{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	ids, err := s.db.GetFollowingIds(ctx, userId.Int64())
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	return ids, nil
}

// returns five random users that people you follow follow, or are in your groups
func (s *Application) GetFollowSuggestions(ctx context.Context, userId ct.Id) ([]models.User, error) {
	input := fmt.Sprintf("%#v", userId)

	if err := userId.Validate(); err != nil {
		return []models.User{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	rows, err := s.db.GetFollowSuggestions(ctx, userId.Int64())
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	users := make([]models.User, 0, len(rows))
	var imageIds ct.Ids
	for _, r := range rows {
		users = append(users, models.User{
			UserId:   ct.Id(r.ID),
			Username: ct.Username(r.Username),
			AvatarId: ct.Id(r.AvatarID),
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
			for i := range users {
				users[i].AvatarURL = avatarMap[users[i].AvatarId.Int64()]
			}
			s.removeFailedImagesAsync(ctx, failedImageIds)
		}
	}

	return users, nil
}

// NOT GRPC
func (s *Application) isFollowRequestPending(ctx context.Context, req models.FollowUserReq) (bool, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return false, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	isPending, err := s.db.IsFollowRequestPending(ctx, ds.IsFollowRequestPendingParams{
		RequesterID: req.FollowerId.Int64(),
		TargetID:    req.TargetUserId.Int64(),
	})
	if err != nil {
		return false, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	return isPending, nil
}

func (s *Application) IsFollowing(ctx context.Context, req models.FollowUserReq) (bool, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return false, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}
	isfollowing, err := s.db.IsFollowing(ctx, ds.IsFollowingParams{
		FollowerID:  req.FollowerId.Int64(),
		FollowingID: req.TargetUserId.Int64(),
	})
	if err != nil {
		return false, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	return isfollowing, nil
}

func (s *Application) AreFollowingEachOther(ctx context.Context, req models.FollowUserReq) (models.FollowRelationship, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return models.FollowRelationship{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	row, err := s.db.AreFollowingEachOther(ctx, ds.AreFollowingEachOtherParams{
		FollowerID:  req.FollowerId.Int64(),
		FollowingID: req.TargetUserId.Int64(),
	})
	if err != nil {
		return models.FollowRelationship{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	followRelationship := models.FollowRelationship{
		FollowerFollowsTarget: row.User1FollowsUser2,
		TargetFollowsFollower: row.User2FollowsUser1,
	}
	tele.Info(ctx, "are following each other: @1", "response", followRelationship)
	return followRelationship, nil //neither follows the other
}

// ---------------------------------------------------------------------
// low priority
// ---------------------------------------------------------------------
func GetMutualFollowers() {}

//get pending follow requests for user
