package application

import (
	"context"
	"fmt"
	ds "social-network/services/posts/internal/db/dbservice"
	"social-network/shared/gen-go/media"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	"social-network/shared/go/models"
	tele "social-network/shared/go/telemetry"

	"github.com/jackc/pgx/v5/pgtype"
)

func (s *Application) GetPersonalizedFeed(ctx context.Context, req models.GetPersonalizedFeedReq) ([]models.Post, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	idsRequesterFollows, err := s.clients.GetFollowingIds(ctx, req.RequesterId.Int64())
	if err != nil {
		return nil, ce.DecodeProto(err, input)
	}

	rows, err := s.db.GetPersonalizedFeed(ctx, ds.GetPersonalizedFeedParams{
		UserID:  req.RequesterId.Int64(),
		Column2: idsRequesterFollows,
		Offset:  req.Offset.Int32(),
		Limit:   req.Limit.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if len(rows) == 0 {
		return []models.Post{}, nil
	}

	posts := make([]models.Post, 0, len(rows))
	userIDs := make(ct.Ids, 0, len(rows))
	postImageIds := make(ct.Ids, 0, len(rows))

	for _, r := range rows {
		uid := r.CreatorID
		userIDs = append(userIDs, ct.Id(uid))

		posts = append(posts, models.Post{

			PostId: ct.Id(r.ID),
			Body:   ct.PostBody(r.PostBody),
			User: models.User{
				UserId: ct.Id(uid),
			},
			CommentsCount:   int(r.CommentsCount),
			ReactionsCount:  int(r.ReactionsCount),
			LastCommentedAt: ct.GenDateTime(r.LastCommentedAt.Time),
			CreatedAt:       ct.GenDateTime(r.CreatedAt.Time),
			UpdatedAt:       ct.GenDateTime(r.UpdatedAt.Time),
			LikedByUser:     r.LikedByUser,
			ImageId:         ct.Id(r.Image),
		})

		if r.Image > 0 {
			postImageIds = append(postImageIds, ct.Id(r.Image))
		}

	}

	userMap, err := s.userRetriever.GetUsers(ctx, userIDs)
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	var imageMap map[int64]string
	var failedImageIds []int64
	if len(postImageIds) > 0 {
		imageMap, failedImageIds, err = s.mediaRetriever.GetImages(ctx, postImageIds, media.FileVariant_MEDIUM)
	}
	if err != nil {
		tele.Error(ctx, "media retriever failed for @1", "request", postImageIds, "error", err.Error()) //log error instead of returning
	} else {

		for i := range posts {
			uid := posts[i].User.UserId
			if u, ok := userMap[uid]; ok {
				posts[i].User = u
			}
			posts[i].ImageUrl = imageMap[posts[i].ImageId.Int64()]
		}
		s.removeFailedImagesAsync(ctx, failedImageIds)
	}

	return posts, nil
}

func (s *Application) GetPublicFeed(ctx context.Context, req models.GenericPaginatedReq) ([]models.Post, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	rows, err := s.db.GetPublicFeed(ctx, ds.GetPublicFeedParams{
		UserID: req.RequesterId.Int64(),
		Offset: req.Offset.Int32(),
		Limit:  req.Limit.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if len(rows) == 0 {
		return []models.Post{}, nil
	}

	posts := make([]models.Post, 0, len(rows))
	userIDs := make(ct.Ids, 0, len(rows))
	postImageIds := make(ct.Ids, 0, len(rows))

	for _, r := range rows {
		uid := r.CreatorID
		userIDs = append(userIDs, ct.Id(uid))

		posts = append(posts, models.Post{
			PostId: ct.Id(r.ID),
			Body:   ct.PostBody(r.PostBody),
			User: models.User{
				UserId: ct.Id(uid),
			},
			CommentsCount:   int(r.CommentsCount),
			ReactionsCount:  int(r.ReactionsCount),
			LastCommentedAt: ct.GenDateTime(r.LastCommentedAt.Time),
			CreatedAt:       ct.GenDateTime(r.CreatedAt.Time),
			UpdatedAt:       ct.GenDateTime(r.UpdatedAt.Time),
			LikedByUser:     r.LikedByUser,
			ImageId:         ct.Id(r.Image),
		})
		if r.Image > 0 {
			postImageIds = append(postImageIds, ct.Id(r.Image))
		}

	}

	userMap, err := s.userRetriever.GetUsers(ctx, userIDs)
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	var imageMap map[int64]string
	var failedImageIds []int64
	if len(postImageIds) > 0 {
		imageMap, failedImageIds, err = s.mediaRetriever.GetImages(ctx, postImageIds, media.FileVariant_MEDIUM)
	}
	if err != nil {
		tele.Error(ctx, "media retriever failed for @1", "request", postImageIds, "error", err.Error()) //log error instead of returning
	} else {

		for i := range posts {
			uid := posts[i].User.UserId
			if u, ok := userMap[uid]; ok {
				posts[i].User = u
			}
			posts[i].ImageUrl = imageMap[posts[i].ImageId.Int64()]
		}
		s.removeFailedImagesAsync(ctx, failedImageIds)
	}

	return posts, nil
}

func (s *Application) GetUserPostsPaginated(ctx context.Context, req models.GetUserPostsReq) ([]models.Post, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	isFollowing, err := s.clients.IsFollowing(ctx, req.RequesterId.Int64(), int64(req.CreatorId))
	if err != nil {
		return nil, ce.DecodeProto(err, input)
	}

	rows, err := s.db.GetUserPostsPaginated(ctx, ds.GetUserPostsPaginatedParams{
		CreatorID: req.CreatorId.Int64(),
		UserID:    req.RequesterId.Int64(),
		Column3:   isFollowing,
		Limit:     req.Limit.Int32(),
		Offset:    req.Offset.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}

	if len(rows) == 0 {
		return []models.Post{}, nil
	}

	posts := make([]models.Post, 0, len(rows))
	userIDs := make(ct.Ids, 0, len(rows))
	postImageIds := make(ct.Ids, 0, len(rows))

	for _, r := range rows {
		uid := r.CreatorID
		userIDs = append(userIDs, ct.Id(uid))

		posts = append(posts, models.Post{
			PostId: ct.Id(r.ID),
			Body:   ct.PostBody(r.PostBody),
			User: models.User{
				UserId: ct.Id(uid),
			},
			CommentsCount:   int(r.CommentsCount),
			ReactionsCount:  int(r.ReactionsCount),
			LastCommentedAt: ct.GenDateTime(r.LastCommentedAt.Time),
			CreatedAt:       ct.GenDateTime(r.CreatedAt.Time),
			UpdatedAt:       ct.GenDateTime(r.UpdatedAt.Time),
			LikedByUser:     r.LikedByUser,
			ImageId:         ct.Id(r.Image),
		})
		if r.Image > 0 {
			postImageIds = append(postImageIds, ct.Id(r.Image))
		}

	}

	userMap, err := s.userRetriever.GetUsers(ctx, userIDs)
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	var imageMap map[int64]string
	var failedImageIds []int64
	if len(postImageIds) > 0 {
		imageMap, failedImageIds, err = s.mediaRetriever.GetImages(ctx, postImageIds, media.FileVariant_MEDIUM)
	}
	if err != nil {
		tele.Error(ctx, "media retriever failed for @1", "request", postImageIds, "error", err.Error()) //log error instead of returning
	} else {

		for i := range posts {
			uid := posts[i].User.UserId
			if u, ok := userMap[uid]; ok {
				posts[i].User = u
			}
			posts[i].ImageUrl = imageMap[posts[i].ImageId.Int64()]
		}
		s.removeFailedImagesAsync(ctx, failedImageIds)
	}

	return posts, nil
}

func (s *Application) GetGroupPostsPaginated(ctx context.Context, req models.GetGroupPostsReq) ([]models.Post, error) {
	input := fmt.Sprintf("%#v", req)

	if err := ct.ValidateStruct(req); err != nil {
		return nil, ce.Wrap(ce.ErrInvalidArgument, err, "request validation failed", input).WithPublic("invalid data received")
	}

	var groupId pgtype.Int8
	groupId.Int64 = req.GroupId.Int64()

	if req.GroupId == 0 {
		return nil, ce.New(ce.ErrInvalidArgument, fmt.Errorf("no group id given"), input).WithPublic("invalid arguments")
	}
	groupId.Valid = true

	isMember, err := s.clients.IsGroupMember(ctx, req.RequesterId.Int64(), req.GroupId.Int64())
	if err != nil {
		return nil, ce.DecodeProto(err, input)
	}
	if !isMember {
		return nil, ce.New(ce.ErrPermissionDenied, fmt.Errorf("user is not group member"), input).WithPublic("permission denied")
	}

	rows, err := s.db.GetGroupPostsPaginated(ctx, ds.GetGroupPostsPaginatedParams{
		GroupID: groupId,
		UserID:  req.RequesterId.Int64(),
		Limit:   req.Limit.Int32(),
		Offset:  req.Offset.Int32(),
	})
	if err != nil {
		return nil, ce.New(ce.ErrInternal, err, input).WithPublic(genericPublic)
	}
	if len(rows) == 0 {
		return []models.Post{}, nil
	}
	posts := make([]models.Post, 0, len(rows))
	userIDs := make(ct.Ids, 0, len(rows))
	postImageIds := make(ct.Ids, 0, len(rows))

	for _, r := range rows {
		uid := r.CreatorID
		userIDs = append(userIDs, ct.Id(uid))

		posts = append(posts, models.Post{
			PostId: ct.Id(r.ID),
			Body:   ct.PostBody(r.PostBody),
			User: models.User{
				UserId: ct.Id(uid),
			},
			GroupId:         req.GroupId,
			Audience:        ct.Audience(r.Audience),
			CommentsCount:   int(r.CommentsCount),
			ReactionsCount:  int(r.ReactionsCount),
			LastCommentedAt: ct.GenDateTime(r.LastCommentedAt.Time),
			CreatedAt:       ct.GenDateTime(r.CreatedAt.Time),
			UpdatedAt:       ct.GenDateTime(r.UpdatedAt.Time),
			LikedByUser:     r.LikedByUser,
			ImageId:         ct.Id(r.Image),
		})

		if r.Image > 0 {
			postImageIds = append(postImageIds, ct.Id(r.Image))
		}
	}

	userMap, err := s.userRetriever.GetUsers(ctx, userIDs)
	if err != nil {
		return nil, ce.Wrap(nil, err, input).WithPublic("error retrieving user's info")
	}

	var imageMap map[int64]string
	var failedImageIds []int64
	if len(postImageIds) > 0 {
		imageMap, failedImageIds, err = s.mediaRetriever.GetImages(ctx, postImageIds, media.FileVariant_MEDIUM)
	}
	if err != nil {
		tele.Error(ctx, "media retriever failed for @1", "request", postImageIds, "error", err.Error()) //log error instead of returning
	} else {

		for i := range posts {
			uid := posts[i].User.UserId
			if u, ok := userMap[uid]; ok {
				posts[i].User = u
			}
			posts[i].ImageUrl = imageMap[posts[i].ImageId.Int64()]
		}
		s.removeFailedImagesAsync(ctx, failedImageIds)
	}

	return posts, nil
}
