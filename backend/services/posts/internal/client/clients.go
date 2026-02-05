package client

import (
	"context"
	"fmt"
	mediapb "social-network/shared/gen-go/media"
	"social-network/shared/gen-go/notifications"
	notifpb "social-network/shared/gen-go/notifications"
	userpb "social-network/shared/gen-go/users"
	"social-network/shared/go/ct"
	"social-network/shared/go/models"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

// Holds connections to clients
type Clients struct {
	UserClient   userpb.UserServiceClient
	MediaClient  mediapb.MediaServiceClient
	NotifsClient notifications.NotificationServiceClient
}

func NewClients(UserClient userpb.UserServiceClient, MediaClient mediapb.MediaServiceClient, NotifsClient notifpb.NotificationServiceClient) *Clients {
	return &Clients{
		UserClient:   UserClient,
		MediaClient:  MediaClient,
		NotifsClient: NotifsClient,
	}
}

func (c *Clients) IsFollowing(ctx context.Context, userId, targetUserId int64) (bool, error) {
	resp, err := c.UserClient.IsFollowing(ctx, &userpb.IsFollowingRequest{
		FollowerId:   userId,
		TargetUserId: targetUserId,
	})
	if err != nil {
		return false, err
	}
	return resp.Value, nil
}

func (c *Clients) IsGroupMember(ctx context.Context, userId, groupId int64) (bool, error) {
	resp, err := c.UserClient.IsGroupMember(ctx, &userpb.GeneralGroupRequest{
		GroupId: groupId,
		UserId:  userId,
	})
	if err != nil {
		return false, err
	}
	return resp.Value, nil
}

// func (c *Clients) GetBatchBasicUserInfo(ctx context.Context, req *cm.UserIds) (*cm.ListUsers, error) {
// 	resp, err := c.UserClient.GetBatchBasicUserInfo(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp, nil
// }

// func (c *Clients) GetBasicUserInfo(ctx context.Context, req *wrapperspb.Int64Value) (*cm.User, error) {
// 	resp, err := c.UserClient.GetBasicUserInfo(ctx, req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return resp, nil
// }

// func (c *Clients) RemoveImages(ctx context.Context, req *userpb.FailedImageIds) error {
// 	_, err := c.UserClient.RemoveImages(ctx, req)
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

func (c *Clients) GetFollowingIds(ctx context.Context, userId int64) ([]int64, error) {
	req := &wrapperspb.Int64Value{Value: userId}

	resp, err := c.UserClient.GetFollowingIds(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp.Values, nil
}

func (c *Clients) GetAllGroupMemberIds(ctx context.Context, groupId int64) ([]int64, error) {

	resp, err := c.UserClient.GetAllGroupMemberIds(ctx, &userpb.IdReq{Id: groupId})
	if err != nil {
		return nil, err
	}
	return resp.Ids, nil
}

func (c *Clients) CreateNewEvent(ctx context.Context, userId, groupId, eventId int64, groupName, eventTitle string) error {
	req := &notifpb.CreateNewEventRequest{
		UserId:     userId,
		GroupId:    groupId,
		EventId:    eventId,
		GroupName:  groupName,
		EventTitle: eventTitle,
	}
	if c.NotifsClient == nil {
		return fmt.Errorf("NotifsClient is nil")
	}
	_, err := c.NotifsClient.CreateNewEvent(ctx, req)
	return err
}

// for comments too?
func (c *Clients) CreatePostLike(ctx context.Context, userId, likerUserId, postId int64, likerUsername string) error {
	req := &notifpb.CreatePostLikeRequest{
		UserId:        userId,
		LikerUserId:   likerUserId,
		PostId:        postId,
		LikerUsername: likerUsername,
		Aggregate:     true,
	}
	if c.NotifsClient == nil {
		return fmt.Errorf("NotifsClient is nil")
	}
	_, err := c.NotifsClient.CreatePostLike(ctx, req)
	return err
}

func (c *Clients) CreatePostComment(ctx context.Context, userId, commenterId, postId int64, commenterUsername, commentContent string) error {
	req := &notifpb.CreatePostCommentRequest{
		UserId:            userId,
		CommenterUserId:   commenterId,
		PostId:            postId,
		CommenterUsername: commenterUsername,
		CommentContent:    commentContent,
		Aggregate:         true,
	}
	if c.NotifsClient == nil {
		return fmt.Errorf("NotifsClient is nil")
	}
	_, err := c.NotifsClient.CreatePostComment(ctx, req)
	return err
}

func (c *Clients) GetGroupBasicInfo(ctx context.Context, groupId int64) (models.Group, error) {
	g, err := c.UserClient.GetGroupBasicInfo(ctx, &userpb.IdReq{Id: groupId})
	if err != nil {
		return models.Group{}, err
	}

	group := models.Group{
		GroupId:          ct.Id(g.GroupId),
		GroupOwnerId:     ct.Id(g.GroupOwnerId),
		GroupTitle:       ct.Title(g.GroupTitle),
		GroupDescription: ct.About(g.GroupDescription),
		GroupImage:       ct.Id(g.GroupImageId),
	}
	return group, nil
}
