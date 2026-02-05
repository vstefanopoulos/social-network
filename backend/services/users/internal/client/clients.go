package client

import (
	"context"
	chatpb "social-network/shared/gen-go/chat"
	"social-network/shared/gen-go/media"
	mediapb "social-network/shared/gen-go/media"
	"social-network/shared/gen-go/notifications"
	"social-network/shared/go/models"
	rds "social-network/shared/go/redis"
	"time"
)

// Holds connections to clients
type Clients struct {
	ChatClient   chatpb.ChatServiceClient
	NotifsClient notifications.NotificationServiceClient
	MediaClient  mediapb.MediaServiceClient
	RedisClient  *rds.RedisClient
}

func NewClients(chatClient chatpb.ChatServiceClient, notifClient notifications.NotificationServiceClient, mediaClient mediapb.MediaServiceClient, redisClient *rds.RedisClient) *Clients {
	c := &Clients{
		ChatClient:   chatClient,
		NotifsClient: notifClient,
		MediaClient:  mediaClient,
		RedisClient:  redisClient,
	}
	return c
}

func (c *Clients) GetImage(ctx context.Context, imageId int64, variant media.FileVariant) (string, error) {
	req := &mediapb.GetImageRequest{
		ImageId: imageId,
		Variant: 1,
	}
	resp, err := c.MediaClient.GetImage(ctx, req)
	if err != nil {
		return "", err
	}
	return resp.DownloadUrl, nil
}

func (c *Clients) GetObj(ctx context.Context, key string, dest any) error {
	return c.RedisClient.GetObj(ctx, key, dest)
}

func (c *Clients) SetObj(ctx context.Context, key string, value any, exp time.Duration) error {
	return c.RedisClient.SetObj(ctx, key, value, exp)
}

func (c *Clients) Del(ctx context.Context, key string) error {
	return c.RedisClient.Del(ctx, key)
}

func (c *Clients) CreateNotification(ctx context.Context, req models.CreateNotificationRequest) error {
	grpcRec := &notifications.CreateNotificationRequest{
		UserId:         req.UserId.Int64(),
		Title:          req.Title,
		Message:        req.Message,
		Type:           req.Type,
		SourceService:  "users",
		SourceEntityId: req.SourceEntityId,
		NeedsAction:    req.NeedsAction,
		Payload:        req.Payload,
		Aggregate:      req.Aggregate,
	}
	_, err := c.NotifsClient.CreateNotification(ctx, grpcRec)
	return err
}

func (c *Clients) CreateFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64, requesterUsername string) error {
	req := &notifications.CreateFollowRequestRequest{
		TargetUserId:      targetUserID,
		RequesterUserId:   requesterUserID,
		RequesterUsername: requesterUsername,
	}
	_, err := c.NotifsClient.CreateFollowRequest(ctx, req)
	return err
}

func (c *Clients) CreateNewFollower(ctx context.Context, targetUserID, followerUserID int64, followerUsername string) error {
	req := &notifications.CreateNewFollowerRequest{
		TargetUserId:     targetUserID,
		FollowerUserId:   followerUserID,
		FollowerUsername: followerUsername,
	}
	_, err := c.NotifsClient.CreateNewFollower(ctx, req)
	return err
}

func (c *Clients) CreateGroupInvite(ctx context.Context, invitedUserId, inviterUserId, groupId int64, groupName, inviterUsername string) error {
	req := &notifications.CreateGroupInviteRequest{
		InvitedUserId:   invitedUserId,
		InviterUserId:   inviterUserId,
		GroupId:         groupId,
		GroupName:       groupName,
		InviterUsername: inviterUsername,
	}
	_, err := c.NotifsClient.CreateGroupInvite(ctx, req)
	return err
}

func (c *Clients) CreateGroupJoinRequest(ctx context.Context, groupOnwerId, requesterId, groupId int64, groupName, requesterUsername string) error {
	req := &notifications.CreateGroupJoinRequestRequest{
		GroupOwnerId:      groupOnwerId,
		RequesterUserId:   requesterId,
		GroupId:           groupId,
		GroupName:         groupName,
		RequesterUsername: requesterUsername,
	}
	_, err := c.NotifsClient.CreateGroupJoinRequest(ctx, req)
	return err
}

func (c *Clients) CreateFollowRequestAccepted(ctx context.Context, requesterId, targetUserId int64, targetUsername string) error {
	req := &notifications.CreateFollowRequestAcceptedRequest{
		RequesterUserId: requesterId,
		TargetUserId:    targetUserId,
		TargetUsername:  targetUsername,
	}
	_, err := c.NotifsClient.CreateFollowRequestAccepted(ctx, req)
	return err
}

func (c *Clients) CreateFollowRequestRejected(ctx context.Context, requesterId, targetUserId int64, targetUsername string) error {
	req := &notifications.CreateFollowRequestRejectedRequest{
		RequesterUserId: requesterId,
		TargetUserId:    targetUserId,
		TargetUsername:  targetUsername,
	}
	_, err := c.NotifsClient.CreateFollowRequestRejected(ctx, req)
	return err
}

func (c *Clients) CreateGroupInviteAccepted(ctx context.Context, invitedUserId, inviterUserId, groupId int64, groupName, invitedUsername string) error {
	req := &notifications.CreateGroupInviteAcceptedRequest{
		InviterUserId:   inviterUserId,
		InvitedUserId:   invitedUserId,
		GroupId:         groupId,
		GroupName:       groupName,
		InvitedUsername: invitedUsername,
	}
	_, err := c.NotifsClient.CreateGroupInviteAccepted(ctx, req)
	return err
}

func (c *Clients) CreateGroupInviteRejected(ctx context.Context, invitedUserId, inviterUserId, groupId int64, groupName, invitedUsername string) error {
	req := &notifications.CreateGroupInviteRejectedRequest{
		InviterUserId:   inviterUserId,
		InvitedUserId:   invitedUserId,
		GroupId:         groupId,
		GroupName:       groupName,
		InvitedUsername: invitedUsername,
	}
	_, err := c.NotifsClient.CreateGroupInviteRejected(ctx, req)
	return err
}

func (c *Clients) CreateGroupJoinRequestAccepted(ctx context.Context, requesterId, groupOwnerId, groupId int64, groupName string) error {
	req := &notifications.CreateGroupJoinRequestAcceptedRequest{
		RequesterUserId: requesterId,
		GroupOwnerId:    groupOwnerId,
		GroupId:         groupId,
		GroupName:       groupName,
	}
	_, err := c.NotifsClient.CreateGroupJoinRequestAccepted(ctx, req)
	return err
}

func (c *Clients) CreateGroupJoinRequestRejected(ctx context.Context, requesterId, groupOwnerId, groupId int64, groupName string) error {
	req := &notifications.CreateGroupJoinRequestRejectedRequest{
		RequesterUserId: requesterId,
		GroupOwnerId:    groupOwnerId,
		GroupId:         groupId,
		GroupName:       groupName,
	}
	_, err := c.NotifsClient.CreateGroupJoinRequestRejected(ctx, req)
	return err
}

// // on successful follow (public profile or accept follow request)
// func (c *Clients) CreatePrivateConversation(ctx context.Context, userId1, userId2 int64) error {
// 	_, err := c.ChatClient.CreatePrivateConversation(ctx, &chatpb.CreatePrivateConvParams{
// 		UserA: userId1,
// 		UserB: userId2,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// // when group is created there's only the owner
// func (c *Clients) CreateGroupConversation(ctx context.Context, groupId, ownerId int64) error {
// 	_, err := c.ChatClient.CreateGroupConversation(ctx, &chatpb.CreateGroupConvParams{
// 		GroupId: groupId,
// 		UserIds: []int64{ownerId},
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (c *Clients) AddMembersToGroupConversation(ctx context.Context, groupId int64, userIds []int64) error {
// 	_, err := c.ChatClient.AddMembersToGroupConversation(ctx, &chatpb.AddMembersToGroupConversationParams{
// 		GroupId: groupId,
// 		UserIds: userIds,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (c *Clients) DeleteConversationByExactMembers(ctx context.Context, userIds []int64) error {
// 	_, err := c.ChatClient.DeleteConversationByExactMembers(ctx, &chatpb.UserIds{
// 		UserIds: userIds,
// 	})
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

//remove members from group conversation?
//delete group conversation on group delete?
