package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ds "social-network/services/posts/internal/db/dbservice"
	"social-network/shared/gen-go/media"
	notifpb "social-network/shared/gen-go/notifications"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Application) CreateEvent(ctx context.Context, req models.CreateEventReq) (int64, error) {
	input := fmt.Sprintf("%#v", req)

	var eventId int64

	if err := ct.ValidateStruct(req); err != nil {
		return 0, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	isMember, err := s.clients.IsGroupMember(ctx, req.CreatorId.Int64(), req.GroupId.Int64())
	if err != nil {
		return 0, ce.DecodeProto(err, input)
	}
	if !isMember {
		return 0, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user is not group member"), input).WithPublic("permission denied")
	}

	// convert date
	eventDate := pgtype.Date{
		Time:  req.EventDate.Time(),
		Valid: true,
	}
	err = s.txRunner.RunTx(ctx, func(q *ds.Queries) error {

		eventId, err = s.db.CreateEvent(ctx, ds.CreateEventParams{
			EventTitle:     req.Title.String(),
			EventBody:      req.Body.String(),
			EventCreatorID: req.CreatorId.Int64(),
			GroupID:        req.GroupId.Int64(),
			EventDate:      eventDate,
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		if req.ImageId != 0 {
			err = q.UpsertImage(ctx, ds.UpsertImageParams{
				ID:       req.ImageId.Int64(),
				ParentID: eventId,
			})
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
		}

		return nil
	})
	if err != nil {
		return 0, ce.Wrap(nil, err)
	}
	//create notification

	group, err := s.clients.GetGroupBasicInfo(ctx, req.GroupId.Int64())
	if err != nil {
		tele.Error(ctx, "Could not get basic group info for id @1 for new event notif: @2", "userId", req.GroupId, "error", err.Error())
	}

	groupMembers, err := s.clients.GetAllGroupMemberIds(ctx, req.GroupId.Int64())
	if err != nil {
		tele.Error(ctx, "Could not get group members ids for group @1 for new event notif: @2", "groupId", req.GroupId, "error", err.Error())
	}
	tele.Debug(ctx, "Group member ids are @1", "groupMemberIds", groupMembers)

	// build the notification event
	event := &notifpb.NotificationEvent{
		EventType: notifpb.EventType_NEW_EVENT_CREATED,
		Payload: &notifpb.NotificationEvent_NewEventCreated{
			NewEventCreated: &notifpb.NewEventCreated{
				UserId:         groupMembers,
				EventCreatorId: int64(req.CreatorId),
				GroupId:        req.GroupId.Int64(),
				EventId:        eventId,
				GroupName:      group.GroupTitle.String(),
				EventTitle:     req.Title.String(),
			},
		},
	}

	if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
		tele.Error(ctx, "failed to send new event notification: @1", "error", err.Error())
	}
	tele.Info(ctx, "new event notification event created")

	return eventId, nil
}

func (s *Application) DeleteEvent(ctx context.Context, req models.GenericReq) error {
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
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to view or edit entity: %v", req.EntityId), input).WithPublic("permission denied")
	}

	rowsAffected, err := s.db.DeleteEvent(ctx, ds.DeleteEventParams{
		ID:             req.EntityId.Int64(),
		EventCreatorID: req.RequesterId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if rowsAffected != 1 {
		return ce.New(ce.ErrNotFound, fmt.Errorf("event %v not found or not owned by user %v", req.EntityId, req.RequesterId), input).WithPublic("not found")
	}

	return nil
}

func (s *Application) EditEvent(ctx context.Context, req models.EditEventReq) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	accessCtx := accessContext{
		requesterId: req.RequesterId.Int64(),
		entityId:    req.EventId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)

	}
	if !hasAccess {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to edit event: %v", req.EventId), input).WithPublic("permission denied")
	}

	return s.txRunner.RunTx(ctx, func(q *ds.Queries) error {
		// convert date
		eventDate := pgtype.Date{
			Time:  req.EventDate.Time(),
			Valid: true,
		}
		rowsAffected, err := q.EditEvent(ctx, ds.EditEventParams{
			EventTitle:     req.Title.String(),
			EventBody:      req.Body.String(),
			EventDate:      eventDate,
			ID:             req.EventId.Int64(),
			EventCreatorID: req.RequesterId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}
		if rowsAffected != 1 {
			return ce.New(ce.ErrNotFound, fmt.Errorf("event %v not found or not owned by user %v", req.EventId, req.RequesterId), input).WithPublic("not found")

		}
		if req.Image > 0 {
			err := q.UpsertImage(ctx, ds.UpsertImageParams{
				ID:       req.Image.Int64(),
				ParentID: req.EventId.Int64(),
			})
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
		}
		if req.DeleteImage {
			rowsAffected, err := q.DeleteImage(ctx, req.EventId.Int64())
			if err != nil {
				return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("event id: %v", req.EventId)).WithPublic(genericPublic)
			}
			if rowsAffected != 1 {
				tele.Warn(ctx, "EditEvent: image to be deleted not found @1", "request", req)
			}
		}
		return nil
	})

}

func (s *Application) GetEventsByGroupId(ctx context.Context, req models.EntityIdPaginatedReq) ([]models.Event, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	isMember, err := s.clients.IsGroupMember(ctx, req.RequesterId.Int64(), req.EntityId.Int64())
	if err != nil {
		return []models.Event{}, ce.DecodeProto(err, input)
	}
	if !isMember {
		return []models.Event{}, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user is not group member"), input).WithPublic("permission denied")
	}

	rows, err := s.db.GetEventsByGroupId(ctx, ds.GetEventsByGroupIdParams{
		GroupID: req.EntityId.Int64(),
		Offset:  req.Offset.Int32(),
		Limit:   req.Limit.Int32(),
		UserID:  req.RequesterId.Int64(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return []models.Event{}, nil
		}
		return nil, ce.Wrap(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	events := make([]models.Event, 0, len(rows))
	userIDs := make(ct.Ids, 0, len(rows))
	eventImageIds := make(ct.Ids, 0, len(rows))

	for _, r := range rows {
		uid := r.EventCreatorID
		userIDs = append(userIDs, ct.Id(uid))
		var ur *bool
		if r.UserResponse.Valid {
			ur = &r.UserResponse.Bool
		}

		events = append(events, models.Event{
			EventId: ct.Id(r.ID),
			Title:   ct.Title(r.EventTitle),
			Body:    ct.EventBody(r.EventBody),
			User: models.User{
				UserId: ct.Id(uid),
			},
			GroupId:       ct.Id(r.GroupID),
			EventDate:     ct.EventDateTime(r.EventDate.Time),
			GoingCount:    int(r.GoingCount),
			NotGoingCount: int(r.NotGoingCount),
			ImageId:       ct.Id(r.Image),
			CreatedAt:     ct.GenDateTime(r.CreatedAt.Time),
			UpdatedAt:     ct.GenDateTime(r.UpdatedAt.Time),
			UserResponse:  ur,
		})
		if r.Image > 0 {
			eventImageIds = append(eventImageIds, ct.Id(r.Image))
		}
	}

	if len(events) == 0 {
		return events, nil
	}

	userMap, err := s.userRetriever.GetUsers(ctx, userIDs)
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	var imageMap map[int64]string
	var failedImageIds []int64
	if len(eventImageIds) > 0 {
		imageMap, failedImageIds, err = s.mediaRetriever.GetImages(ctx, eventImageIds, media.FileVariant_MEDIUM)
	}
	if err != nil {
		tele.Error(ctx, "media retriever failed for @1", "request", eventImageIds, "error", err.Error()) //log error instead of returning
	} else {
		for i := range events {
			uid := events[i].User.UserId
			if u, ok := userMap[uid]; ok {
				events[i].User = u
			}
			events[i].ImageUrl = imageMap[events[i].ImageId.Int64()]
		}
		s.removeFailedImagesAsync(ctx, failedImageIds)
	}

	return events, nil
}

func (s *Application) RespondToEvent(ctx context.Context, req models.RespondToEventReq) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	accessCtx := accessContext{
		requesterId: req.ResponderId.Int64(),
		entityId:    req.EventId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)
	}
	if !hasAccess {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to respond to event %v", req.EventId), input).WithPublic("permission denied")
	}

	_, err = s.db.UpsertEventResponse(ctx, ds.UpsertEventResponseParams{
		EventID: req.EventId.Int64(),
		UserID:  req.ResponderId.Int64(),
		Going:   req.Going,
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	// if rowsAffected != 1 {
	// 	return ErrNotFound
	// }
	return nil
}

func (s *Application) RemoveEventResponse(ctx context.Context, req models.GenericReq) error {
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
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to remove response with id %v", req.EntityId), input).WithPublic("permission denied")
	}

	rowsAffected, err := s.db.DeleteEventResponse(ctx, ds.DeleteEventResponseParams{
		EventID: req.EntityId.Int64(),
		UserID:  req.RequesterId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if rowsAffected != 1 {
		return ce.New(ce.ErrNotFound, fmt.Errorf("event response %v not found or not owned by user %v", req.EntityId, req.RequesterId), input).WithPublic("not found")
	}
	return nil
}
