package application

import (
	"context"
	"fmt"
	ds "social-network/services/posts/internal/db/dbservice"
	notifpb "social-network/shared/gen-go/notifications"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
)

func (s *Application) ToggleOrInsertReaction(ctx context.Context, req models.GenericReq) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")

	}

	accessCtx := accessContext{
		requesterId: req.RequesterId.Int64(),
		entityId:    req.EntityId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)
	}
	if !hasAccess {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to react to entity %v", req.EntityId), input).WithPublic("permission denied")
	}

	res, err := s.db.ToggleOrInsertReaction(ctx, ds.ToggleOrInsertReactionParams{
		ContentID: req.EntityId.Int64(),
		UserID:    req.RequesterId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if res.ShouldNotify {
		//create notification
		liker, err := s.userRetriever.GetUser(ctx, ct.Id(req.RequesterId))
		if err != nil {

		}

		row, err := s.db.GetEntityCreatorAndGroup(ctx, req.EntityId.Int64())
		if err != nil {
			tele.Error(ctx, "Could not get entity creator and group for toggle or insert reaction for entity @1", "entityId", req.EntityId)
			return nil
		}

		//if liker is same as entity creator, return without creating a notification
		if row.CreatorID == int64(liker.UserId) {
			return nil
		}

		// build the notification event
		event := &notifpb.NotificationEvent{
			EventType: notifpb.EventType_POST_LIKED,
			Payload: &notifpb.NotificationEvent_PostLiked{
				PostLiked: &notifpb.PostLiked{
					EntityCreatorId: row.CreatorID,
					PostId:          req.EntityId.Int64(),
					LikerUserId:     req.RequesterId.Int64(),
					LikerUsername:   liker.Username.String(),
					Aggregate:       true,
				},
			},
		}

		if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
			tele.Error(ctx, "failed to send new reaction notification: @1", "error", err.Error())
		}
		tele.Info(ctx, "new reaction notification event created")

	}
	return nil
}

func (s *Application) GetWhoLikedEntityId(ctx context.Context, req models.GenericReq) ([]models.User, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")

	}

	userIDs, err := s.db.GetWhoLikedEntityId(ctx, req.EntityId.Int64())
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if len(userIDs) < 1 {
		return []models.User{}, nil
	}

	userMap, err := s.userRetriever.GetUsers(ctx, ct.FromInt64s(userIDs))
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	users := make([]models.User, 0, len(userIDs))
	for _, id := range ct.FromInt64s(userIDs) {
		if u, ok := userMap[id]; ok {
			users = append(users, u)
		}
	}

	return users, nil
}
