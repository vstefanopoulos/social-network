package dbservice

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
)

type Querier interface {
	CanUserSeeEntity(ctx context.Context, arg CanUserSeeEntityParams) (bool, error)
	ClearPostAudience(ctx context.Context, postID int64) error
	CreateComment(ctx context.Context, arg CreateCommentParams) (int64, error)
	CreateEvent(ctx context.Context, arg CreateEventParams) (int64, error)
	CreatePost(ctx context.Context, arg CreatePostParams) (int64, error)
	DeleteComment(ctx context.Context, arg DeleteCommentParams) (int64, error)
	DeleteEvent(ctx context.Context, arg DeleteEventParams) (int64, error)
	DeleteEventResponse(ctx context.Context, arg DeleteEventResponseParams) (int64, error)
	DeleteImage(ctx context.Context, id int64) (int64, error)
	DeletePost(ctx context.Context, arg DeletePostParams) (int64, error)
	EditComment(ctx context.Context, arg EditCommentParams) (int64, error)
	EditEvent(ctx context.Context, arg EditEventParams) (int64, error)
	EditPostContent(ctx context.Context, arg EditPostContentParams) (int64, error)
	GetBasicPostByID(ctx context.Context, postId int64) (GetBasicPostByIDRow, error)
	GetCommentsByPostId(ctx context.Context, arg GetCommentsByPostIdParams) ([]GetCommentsByPostIdRow, error)
	GetEntityCreatorAndGroup(ctx context.Context, id int64) (GetEntityCreatorAndGroupRow, error)
	GetEventsByGroupId(ctx context.Context, arg GetEventsByGroupIdParams) ([]GetEventsByGroupIdRow, error)
	GetGroupPostsPaginated(ctx context.Context, arg GetGroupPostsPaginatedParams) ([]GetGroupPostsPaginatedRow, error)
	GetImages(ctx context.Context, parentID int64) (int64, error)
	GetLatestCommentforPostId(ctx context.Context, arg GetLatestCommentforPostIdParams) (GetLatestCommentforPostIdRow, error)
	GetMostPopularPostInGroup(ctx context.Context, groupID pgtype.Int8) (GetMostPopularPostInGroupRow, error)
	GetPostAudienceForComment(ctx context.Context, postID int64) (string, error)
	GetPersonalizedFeed(ctx context.Context, arg GetPersonalizedFeedParams) ([]GetPersonalizedFeedRow, error)
	GetPostAudience(ctx context.Context, postID int64) ([]int64, error)
	GetPostByID(ctx context.Context, arg GetPostByIDParams) (GetPostByIDRow, error)
	GetPublicFeed(ctx context.Context, arg GetPublicFeedParams) ([]GetPublicFeedRow, error)
	// pagination
	GetUserPostsPaginated(ctx context.Context, arg GetUserPostsPaginatedParams) ([]GetUserPostsPaginatedRow, error)
	GetWhoLikedEntityId(ctx context.Context, contentID int64) ([]int64, error)
	InsertPostAudience(ctx context.Context, arg InsertPostAudienceParams) (int64, error)
	// U1: Users who liked one or more of *your public posts*
	// U2: Users who commented on your public posts
	// U3: Users who liked the same posts as you
	// U4: Users who commented on the same posts as you
	// Combine scores
	SuggestUsersByPostActivity(ctx context.Context, creatorID int64) ([]int64, error)
	ToggleOrInsertReaction(ctx context.Context, arg ToggleOrInsertReactionParams) (ToggleOrInsertReactionResult, error)
	UpdatePostAudience(ctx context.Context, arg UpdatePostAudienceParams) (int64, error)
	UpsertEventResponse(ctx context.Context, arg UpsertEventResponseParams) (int64, error)
	UpsertImage(ctx context.Context, arg UpsertImageParams) error
}

var _ Querier = (*Queries)(nil)
