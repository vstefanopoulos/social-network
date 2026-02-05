package application

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	ds "social-network/services/posts/internal/db/dbservice"
	"social-network/shared/gen-go/media"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Application) CreatePost(ctx context.Context, req models.CreatePostReq) (postId int64, err error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return 0, ce.Wrap(ce.ErrInvalidArgument, err, input).WithPublic("invalid data received")
	}

	var groupId pgtype.Int8
	groupId.Int64 = req.GroupId.Int64()
	if req.GroupId == 0 {
		groupId.Valid = false
	} else {
		groupId.Valid = true
	}

	audience := ds.IntendedAudience(req.Audience.String())

	if !groupId.Valid && audience == "group" {
		return 0, ce.New(ce.ErrInvalidArgument, fmt.Errorf("no group id given"), input).WithPublic("invalid arguments")
	}

	if groupId.Valid {
		isMember, err := s.clients.IsGroupMember(ctx, req.CreatorId.Int64(), req.GroupId.Int64())
		if err != nil {
			return 0, ce.DecodeProto(err, input)
		}
		if !isMember {
			return 0, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user is not a member of group %v", req.GroupId), input).WithPublic("permission denied")

		}
	}
	err = s.txRunner.RunTx(ctx, func(q *ds.Queries) error {

		postId, err = q.CreatePost(ctx, ds.CreatePostParams{
			PostBody:  req.Body.String(),
			CreatorID: req.CreatorId.Int64(),
			GroupID:   groupId,
			Audience:  audience,
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}

		if audience == "selected" {
			audienceIds := ct.Ids(req.AudienceIds).Unique()
			if len(audienceIds) < 1 {
				return ce.New(ce.ErrInvalidArgument, fmt.Errorf("no audience given for post with audience=selected"), input).WithPublic(genericPublic)
			}
			rowsAffected, err := q.InsertPostAudience(ctx, ds.InsertPostAudienceParams{
				PostID:         postId,
				AllowedUserIds: audienceIds.Int64(),
			})
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
			if rowsAffected < int64(len(audienceIds)) {
				return ce.New(ce.ErrInternal, fmt.Errorf("unexpected rows returned: expected %v, got %v", len(audienceIds), rowsAffected), input).WithPublic(genericPublic)
			}
		}

		if req.ImageId != 0 {
			err = q.UpsertImage(ctx, ds.UpsertImageParams{
				ID:       req.ImageId.Int64(),
				ParentID: postId,
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
	return postId, nil
}

func (s *Application) DeletePost(ctx context.Context, req models.GenericReq) error {
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

	rowsAffected, err := s.db.DeletePost(ctx, ds.DeletePostParams{
		ID:        int64(req.EntityId),
		CreatorID: req.RequesterId.Int64(),
	})
	if err != nil {
		return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if rowsAffected != 1 {
		return ce.New(ce.ErrNotFound, fmt.Errorf("post %v not found or not owned by user %v", req.EntityId, req.RequesterId), input).WithPublic("not found")
	}
	return nil
}

func (s *Application) EditPost(ctx context.Context, req models.EditPostReq) error {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	accessCtx := accessContext{
		requesterId: req.RequesterId.Int64(),
		entityId:    req.PostId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)
	}
	if !hasAccess {
		return ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to view or edit entity %v", req.PostId), input).WithPublic("permission denied")
	}

	return s.txRunner.RunTx(ctx, func(q *ds.Queries) error {
		//edit content
		if len(req.NewBody) > 0 {
			rowsAffected, err := q.EditPostContent(ctx, ds.EditPostContentParams{
				PostBody:  req.NewBody.String(),
				ID:        req.PostId.Int64(),
				CreatorID: req.RequesterId.Int64(),
			})
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
			if rowsAffected != 1 {
				return ce.New(ce.ErrNotFound, fmt.Errorf("post %v not found or not owned by user %v", req.PostId, req.RequesterId), input).WithPublic("not found")
			}
		}

		//edit image
		if req.ImageId > 0 {
			err := q.UpsertImage(ctx, ds.UpsertImageParams{
				ID:       req.ImageId.Int64(),
				ParentID: req.PostId.Int64(),
			})
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
		}
		//delete image
		if req.DeleteImage {
			rowsAffected, err := q.DeleteImage(ctx, req.PostId.Int64())
			if err != nil {
				return ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("post id: %v", req.PostId)).WithPublic(genericPublic)
			}
			if rowsAffected != 1 {
				tele.Warn(ctx, "image @1 for post @2 could not be deleted: not found.", "image id", req.ImageId, "post id", req.PostId)
			}
		}
		// edit audience
		_, err := q.UpdatePostAudience(ctx, ds.UpdatePostAudienceParams{
			ID:        req.PostId.Int64(),
			CreatorID: req.RequesterId.Int64(),
			Audience:  ds.IntendedAudience(req.Audience),
		})
		if err != nil {
			return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
		}
		// if rowsAffected != 1 {
		// 	tele.Warn(ctx, "EditPost: no audience change. @1", "request", req)
		// }

		// edit audience ids
		if req.Audience == "selected" {
			audienceIds := ct.Ids(req.AudienceIds).Unique()
			if len(audienceIds) < 1 {
				return ce.New(ce.ErrInvalidArgument, fmt.Errorf("no audience given for post with audience=selected"), input).WithPublic(genericPublic)
			}
			//delete previous audience ids
			err := q.ClearPostAudience(ctx, req.PostId.Int64())
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
			//insert new ids
			rowsAffected, err := q.InsertPostAudience(ctx, ds.InsertPostAudienceParams{
				PostID:         req.PostId.Int64(),
				AllowedUserIds: audienceIds.Int64(),
			})
			if err != nil {
				return ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
			}
			if rowsAffected < int64(len(audienceIds)) {
				return ce.New(ce.ErrInternal, fmt.Errorf("unexpected rows returned: expected %v, got %v", len(audienceIds), rowsAffected), input).WithPublic(genericPublic)
			}
		}

		return nil
	})

}

func (s *Application) GetMostPopularPostInGroup(ctx context.Context, req models.SimpleIdReq) (models.Post, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return models.Post{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	var groupId pgtype.Int8
	groupId.Int64 = req.Id.Int64()
	if req.Id == 0 {
		return models.Post{}, ce.New(ce.ErrInvalidArgument, fmt.Errorf("no group id given"), input).WithPublic("invalid arguments")
	}
	groupId.Valid = true

	p, err := s.db.GetMostPopularPostInGroup(ctx, groupId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Post{}, nil
		}
		return models.Post{}, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	userMap, err := s.userRetriever.GetUsers(ctx, ct.Ids{ct.Id(p.CreatorID)})
	if err != nil {
		return models.Post{}, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	post := models.Post{
		PostId:          ct.Id(p.ID),
		Body:            ct.PostBody(p.PostBody),
		User:            userMap[ct.Id(p.CreatorID)],
		GroupId:         ct.Id(req.Id.Int64()),
		Audience:        ct.Audience(p.Audience),
		CommentsCount:   int(p.CommentsCount),
		ReactionsCount:  int(p.ReactionsCount),
		LastCommentedAt: ct.GenDateTime(p.LastCommentedAt.Time),
		CreatedAt:       ct.GenDateTime(p.CreatedAt.Time),
		UpdatedAt:       ct.GenDateTime(p.UpdatedAt.Time),
		ImageId:         ct.Id(p.Image),
	}

	if post.ImageId > 0 {
		imageUrl, err := s.mediaRetriever.GetImage(ctx, p.Image, media.FileVariant_MEDIUM)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", p.Image, "error", err.Error()) //log error instead of returning
			s.removeFailedImageAsync(ctx, err, p.Image)
		} else {

			post.ImageUrl = imageUrl
		}
	}

	return post, nil
}

func (s *Application) GetPostById(ctx context.Context, req models.GenericReq) (models.Post, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return models.Post{}, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	accessCtx := accessContext{
		requesterId: req.RequesterId.Int64(),
		entityId:    req.EntityId.Int64(),
	}

	hasAccess, err := s.hasRightToView(ctx, accessCtx)
	if err != nil {
		return models.Post{}, ce.Wrap(ce.ErrInternal, err, fmt.Sprintf("%#v", accessCtx)).WithPublic(genericPublic)
	}
	if !hasAccess {
		return models.Post{}, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user has no permission to view or edit entity %v", req.EntityId), input).WithPublic("permission denied")
	}

	p, err := s.db.GetPostByID(ctx, ds.GetPostByIDParams{
		UserID: req.RequesterId.Int64(),
		ID:     req.EntityId.Int64(),
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.Post{}, ce.New(ce.ErrNotFound, err, input).WithPublic("not found")
		}
		return models.Post{}, err
	}
	var userIds ct.Ids
	//add creator
	userIds = append(userIds, ct.Id(p.CreatorID))
	//add selected audience if any
	if len(p.SelectedAudience) > 0 {
		for _, id := range p.SelectedAudience {
			if id > 0 {
				userIds = append(userIds, ct.Id(id))
			}
		}
	}

	userMap, err := s.userRetriever.GetUsers(ctx, userIds.Unique())
	if err != nil {
		return models.Post{}, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	selectedUsers := make([]models.User, 0, len(p.SelectedAudience))

	if len(p.SelectedAudience) > 0 {
		for _, id := range p.SelectedAudience {
			selectedUsers = append(selectedUsers, userMap[ct.Id(id)])
		}
	}

	post := models.Post{
		PostId:                ct.Id(p.ID),
		Body:                  ct.PostBody(p.PostBody),
		User:                  userMap[ct.Id(p.CreatorID)],
		GroupId:               ct.Id(p.GroupID),
		Audience:              ct.Audience(p.Audience),
		CommentsCount:         int(p.CommentsCount),
		ReactionsCount:        int(p.ReactionsCount),
		LastCommentedAt:       ct.GenDateTime(p.LastCommentedAt.Time),
		CreatedAt:             ct.GenDateTime(p.CreatedAt.Time),
		UpdatedAt:             ct.GenDateTime(p.UpdatedAt.Time),
		LikedByUser:           p.LikedByUser,
		ImageId:               ct.Id(p.Image),
		SelectedAudienceUsers: selectedUsers,
	}

	if post.ImageId > 0 {
		imageUrl, err := s.mediaRetriever.GetImage(ctx, p.Image, media.FileVariant_MEDIUM)
		if err != nil {
			tele.Error(ctx, "media retriever failed for @1", "request", p.Image, "error", err.Error()) //log error instead of returning
			s.removeFailedImageAsync(ctx, err, p.Image)
		} else {

			post.ImageUrl = imageUrl
		}
	}

	return post, nil
}
