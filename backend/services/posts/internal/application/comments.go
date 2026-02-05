package application

import (
	"context"
	"fmt"
	ds "social-network/services/posts/internal/db/dbservice"
	"social-network/shared/gen-go/media"
	notifpb "social-network/shared/gen-go/notifications"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"
)

const genericPublic = "posts service error"

func (s *Application) CreateComment(ctx context.Context, req models.CreateCommentReq) (commentId int64, err error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return 0, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	accessCtx := accessContext{
		requesterId: req.CreatorId.Int64(),
		entityId:    req.ParentId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return 0, ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)
	}
	if !hasAccess {
		return 0, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to comment on post: %v", req.ParentId), input).WithPublic("permission denied")
	}

	err = s.txRunner.RunTx(ctx, func(q *ds.Queries) error {
		commentId, err = q.CreateComment(ctx, ds.CreateCommentParams{
			CommentCreatorID: req.CreatorId.Int64(),
			ParentID:         req.ParentId.Int64(),
			CommentBody:      req.Body.String(),
		})

		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		if req.ImageId != 0 {
			err = q.UpsertImage(ctx, ds.UpsertImageParams{
				ID:       req.ImageId.Int64(),
				ParentID: commentId,
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
	commenter, err := s.userRetriever.GetUser(ctx, ct.Id(req.CreatorId))
	if err != nil {
		tele.Error(ctx, "Could not get basic user info for id @1 for comment created event: @2", "userId", req.CreatorId, "error", err.Error())
	}

	basicPost, err := s.db.GetBasicPostByID(ctx, req.ParentId.Int64())
	if err != nil {
		tele.Error(ctx, "Could not get basic post info for id @1 for comment created event: @2", "postId", req.ParentId, "error", err.Error())
	}

	//if commenter is parent creator, do not create notification
	if commenter.UserId == ct.Id(basicPost.CreatorID) {
		return commentId, nil
	}

	// build the notification event
	event := &notifpb.NotificationEvent{
		EventType: notifpb.EventType_POST_COMMENT_CREATED,
		Payload: &notifpb.NotificationEvent_PostCommentCreated{
			PostCommentCreated: &notifpb.PostCommentCreated{
				PostCreatorId:     basicPost.CreatorID,
				PostId:            req.ParentId.Int64(),
				CommentId:         commentId,
				CommenterUserId:   req.CreatorId.Int64(),
				CommenterUsername: commenter.Username.String(),
				Body:              basicPost.PostBody,
				Aggregate:         true,
			},
		},
	}

	if err := s.eventProducer.CreateAndSendNotificationEvent(ctx, event); err != nil {
		tele.Error(ctx, "failed to send new comment notification: @1", "error", err.Error())
	}
	tele.Info(ctx, "new comment notification event created")

	return commentId, nil
}

func (s *Application) EditComment(ctx context.Context, req models.EditCommentReq) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	accessCtx := accessContext{
		requesterId: req.CreatorId.Int64(),
		entityId:    req.CommentId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)
	}
	if !hasAccess {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to view or edit entity %v", req.CommentId), input).WithPublic("permission denied")
	}

	err = s.txRunner.RunTx(ctx, func(q *ds.Queries) error {
		rowsAffected, err := q.EditComment(ctx, ds.EditCommentParams{
			CommentBody:      req.Body.String(),
			ID:               req.CommentId.Int64(),
			CommentCreatorID: req.CreatorId.Int64(),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}
		if rowsAffected != 1 {
			return ce.New(ce.ErrNotFound, fmt.Errorf("comment %v not found or not owned by user %v", req.CommentId, req.CreatorId), input).WithPublic("not found")
		}

		if req.ImageId > 0 {
			err := q.UpsertImage(ctx, ds.UpsertImageParams{
				ID:       req.ImageId.Int64(),
				ParentID: req.CommentId.Int64(),
			})
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
		}

		if req.DeleteImage {
			rowsAffected, err := q.DeleteImage(ctx, req.CommentId.Int64())
			if err != nil {
				return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("comment id: %v", req.CommentId)).WithPublic(genericPublic)
			}
			if rowsAffected != 1 {
				tele.Warn(ctx, "image @1 for comment @2 could not be deleted: not found.", "imageId", req.ImageId, "commenId", req.CommentId)
			}
		}
		return nil
	})
	if err != nil {
		return ce.Wrap(nil, err)
	}

	return nil
}
func (s *Application) DeleteComment(ctx context.Context, req models.GenericReq) error {
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
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to delete entity %v", req.EntityId), input).WithPublic("permission denied")
	}

	rowsAffected, err := s.db.DeleteComment(ctx, ds.DeleteCommentParams{
		ID:               req.EntityId.Int64(),
		CommentCreatorID: req.RequesterId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if rowsAffected != 1 {
		return ce.New(ce.ErrNotFound, fmt.Errorf("comment %v not found or not owned by user %v", req.EntityId, req.RequesterId), input).WithPublic("not found")
	}

	return nil
}

func (s *Application) GetCommentsByParentId(ctx context.Context, req models.EntityIdPaginatedReq) ([]models.Comment, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	accessCtx := accessContext{
		requesterId: req.RequesterId.Int64(),
		entityId:    req.EntityId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return nil, ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)
	}
	if !hasAccess {
		return nil, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to view comments of entity %v", req.EntityId), input).WithPublic("permission denied")
	}

	rows, err := s.db.GetCommentsByPostId(ctx, ds.GetCommentsByPostIdParams{
		ParentID: req.EntityId.Int64(),
		UserID:   req.RequesterId.Int64(),
		Limit:    req.Limit.Int32(),
		Offset:   req.Offset.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if len(rows) == 0 {
		return []models.Comment{}, nil
	}

	comments := make([]models.Comment, 0, len(rows))
	userIDs := make(ct.Ids, 0, len(rows))
	commentImageIds := make(ct.Ids, 0, len(rows))

	for _, r := range rows {
		userIDs = append(userIDs, ct.Id(r.CommentCreatorID))

		comments = append(comments, models.Comment{
			CommentId: ct.Id(r.ID),
			ParentId:  req.EntityId,
			Body:      ct.CommentBody(r.CommentBody),
			User: models.User{
				UserId: ct.Id(r.CommentCreatorID),
			},
			ReactionsCount: int(r.ReactionsCount),
			CreatedAt:      ct.GenDateTime(r.CreatedAt.Time),
			UpdatedAt:      ct.GenDateTime(r.UpdatedAt.Time),
			LikedByUser:    r.LikedByUser,
			ImageId:        ct.Id(r.Image),
		})

		if r.Image > 0 {
			commentImageIds = append(commentImageIds, ct.Id(r.Image))
		}
	}

	userMap, err := s.userRetriever.GetUsers(ctx, userIDs)
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	var imageMap map[int64]string
	var failedImageIds []int64
	if len(commentImageIds) > 0 {
		imageMap, failedImageIds, err = s.mediaRetriever.GetImages(ctx, commentImageIds, media.FileVariant_MEDIUM)
	}
	if err != nil {
		tele.Error(ctx, "media retriever failed for @1", "request", commentImageIds, "error", err.Error()) //log error instead of returning
	} else {

		for i := range comments {
			uid := comments[i].User.UserId
			if u, ok := userMap[uid]; ok {
				comments[i].User = u
			}
			comments[i].ImageUrl = imageMap[comments[i].ImageId.Int64()]
		}
		s.removeFailedImagesAsync(ctx, failedImageIds)
	}

	return comments, nil
}

func (s *Application) GetPostAudienceForComment(ctx context.Context, postId int64) (string, error) {
	input := fmt.Sprintf("commentId: %v", postId)
	audience, err := s.db.GetPostAudienceForComment(ctx, postId)
	if err != nil {
		return "", ce.Wrap(nil, err, input).WithPublic("error retrieving comment's visibility")
	}
	return audience, nil
}
