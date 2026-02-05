package handler

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"

	"social-network/services/notifications/internal/application"
	pb "social-network/shared/gen-go/notifications"
	"social-network/shared/go/ct"
)

// CreateNotification creates a new notification
func (s *Server) CreateNotification(ctx context.Context, req *pb.CreateNotificationRequest) (*pb.Notification, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	payload := make(map[string]string)
	for k, v := range req.Payload {
		payload[k] = v
	}

	// For now, use aggregation based on notification type (like, comment, follower, message)
	// When protobuf is regenerated with the Aggregate field, this can be controlled from the request
	aggregate := shouldAggregateNotification(req.Type)

	notification, err := s.Application.CreateNotificationWithAggregation(
		ctx,
		req.UserId,
		convertProtoNotificationTypeToApplication(req.Type),
		req.Title,
		req.Message,
		req.SourceService,
		req.SourceEntityId,
		req.NeedsAction,
		payload,
		aggregate,
	)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create notification: %v", err)
	}

	return s.convertToProtoNotification(notification), nil
}

// CreateNotifications creates multiple notifications
func (s *Server) CreateNotifications(ctx context.Context, req *pb.CreateNotificationsRequest) (*pb.CreateNotificationsResponse, error) {
	// Create notifications individually to allow for aggregation control
	createdNotifications := make([]*application.Notification, 0, len(req.Notifications))

	for _, n := range req.Notifications {
		payload := make(map[string]string)
		for k, v := range n.Payload {
			payload[k] = v
		}

		// For now, use aggregation based on notification type (like, comment, follower, message)
		// When protobuf is regenerated with the Aggregate field, this can be controlled from the request
		aggregate := shouldAggregateNotification(n.Type)

		notification, err := s.Application.CreateNotificationWithAggregation(
			ctx,
			n.UserId,
			convertProtoNotificationTypeToApplication(n.Type),
			n.Title,
			n.Message,
			n.SourceService,
			n.SourceEntityId,
			n.NeedsAction,
			payload,
			aggregate, // Use the aggregate flag determined by type
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to create notification: %v", err)
		}

		createdNotifications = append(createdNotifications, notification)
	}

	pbNotifications := make([]*pb.Notification, len(createdNotifications))
	for i, n := range createdNotifications {
		pbNotifications[i] = s.convertToProtoNotification(n)
	}

	return &pb.CreateNotificationsResponse{
		CreatedNotifications: pbNotifications,
	}, nil
}

// GetUserNotifications retrieves notifications for a user -0
func (s *Server) GetUserNotifications(ctx context.Context, req *pb.GetUserNotificationsRequest) (*pb.GetUserNotificationsResponse, error) {
	if req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	if req.Limit == 0 {
		req.Limit = 20 // default limit
	}

	notifications, err := s.Application.GetUserNotifications(ctx, req.UserId, req.Limit, req.Offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user notifications: %v", err)
	}

	// Convert to protobuf notifications
	pbNotifications := make([]*pb.Notification, len(notifications))
	for i, n := range notifications {
		pbNotifications[i] = s.convertToProtoNotification(n)
	}

	// Get total count
	totalCount, err := s.Application.GetUserNotificationsCount(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get notifications count: %v", err)
	}

	// Get unread count
	unreadCount, err := s.Application.GetUserUnreadNotificationsCount(ctx, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get unread notifications count: %v", err)
	}

	return &pb.GetUserNotificationsResponse{
		Notifications: pbNotifications,
		TotalCount:    int32(totalCount),
		UnreadCount:   int32(unreadCount),
	}, nil
}

// GetUnreadNotificationsCount returns the count of unread notifications for a user -0
func (s *Server) GetUnreadNotificationsCount(ctx context.Context, req *wrapperspb.Int64Value) (*wrapperspb.Int64Value, error) {
	if req.Value == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	count, err := s.Application.GetUserUnreadNotificationsCount(ctx, req.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get unread notifications count: %v", err)
	}

	return &wrapperspb.Int64Value{
		Value: count,
	}, nil
}

// MarkNotificationAsRead marks a notification as read -0
func (s *Server) MarkNotificationAsRead(ctx context.Context, req *pb.MarkNotificationAsReadRequest) (*emptypb.Empty, error) {
	if req.NotificationId == 0 || req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "notification_id and user_id are required")
	}

	err := s.Application.MarkNotificationAsRead(ctx, req.NotificationId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark notification as read: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// MarkNotificationAsActed marks a notification as acted -0
func (s *Server) MarkNotificationAsActed(ctx context.Context, req *pb.MarkNotificationAsActedRequest) (*emptypb.Empty, error) {
	if req.NotificationId == 0 || req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "notification_id and user_id are required")
	}

	err := s.Application.MarkNotificationAsActed(ctx, req.NotificationId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark notification as acted: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// MarkAllAsRead marks all notifications for a user as read -0
func (s *Server) MarkAllAsRead(ctx context.Context, req *wrapperspb.Int64Value) (*emptypb.Empty, error) {
	if req.Value == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	err := s.Application.MarkAllAsRead(ctx, req.Value)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to mark all notifications as read: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// DeleteNotification deletes a notification -0
func (s *Server) DeleteNotification(ctx context.Context, req *pb.DeleteNotificationRequest) (*emptypb.Empty, error) {
	if req.NotificationId == 0 || req.UserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "notification_id and user_id are required")
	}

	err := s.Application.DeleteNotification(ctx, req.NotificationId, req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete notification: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// CreateFollowRequest creates a follow request notification
func (s *Server) CreateFollowRequest(ctx context.Context, req *pb.CreateFollowRequestRequest) (*pb.Notification, error) {
	if req.TargetUserId == 0 || req.RequesterUserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "target_user_id and requester_user_id are required")
	}

	err := s.Application.CreateFollowRequestNotification(ctx, req.TargetUserId, req.RequesterUserId, req.RequesterUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create follow request notification: %v", err)
	}

	// Return a success response by fetching the notification that was created
	// Since CreateFollowRequestNotification doesn't return the notification, we need to create it manually
	// or return an empty notification with success status
	// For now, we return an empty notification as the original function returns only error
	// To properly return the notification, we'd need to modify the internal function
	// For consistency with existing functions, we'll return a basic notification

	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.FollowRequest,
		Title:          "New Follow Request",
		Message:        fmt.Sprintf("%s wants to follow you", req.RequesterUsername),
		SourceEntityID: ct.Id(req.RequesterUserId),
		SourceService:  "users",
		NeedsAction:    true,
		Payload: map[string]string{
			"requester_id":   fmt.Sprintf("%d", req.RequesterUserId),
			"requester_name": req.RequesterUsername,
		},
		// Set other fields as needed
	})

	return notification, nil
}

// CreateNewFollower creates a new follower notification
func (s *Server) CreateNewFollower(ctx context.Context, req *pb.CreateNewFollowerRequest) (*pb.Notification, error) {
	if req.TargetUserId == 0 || req.FollowerUserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "target_user_id and follower_user_id are required")
	}

	err := s.Application.CreateNewFollowerNotification(ctx, req.TargetUserId, req.FollowerUserId, req.FollowerUsername, req.Aggregate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create new follower notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.NewFollower,
		Title:          "New Follower",
		Message:        fmt.Sprintf("%s is now following you", req.FollowerUsername),
		SourceEntityID: ct.Id(req.FollowerUserId),
		SourceService:  "users",
		NeedsAction:    false,
		Payload: map[string]string{
			"follower_id":   fmt.Sprintf("%d", req.FollowerUserId),
			"follower_name": req.FollowerUsername,
		},
	})

	return notification, nil
}

// CreateGroupInvite creates a group invite notification
func (s *Server) CreateGroupInvite(ctx context.Context, req *pb.CreateGroupInviteRequest) (*pb.Notification, error) {
	if req.InvitedUserId == 0 || req.InviterUserId == 0 || req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invited_user_id, inviter_user_id, and group_id are required")
	}

	err := s.Application.CreateGroupInviteNotification(ctx, req.InvitedUserId, req.InviterUserId, req.GroupId, req.GroupName, req.InviterUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group invite notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.GroupInvite,
		Title:          "Group Invitation",
		Message:        fmt.Sprintf("%s invited you to join the group \"%s\"", req.InviterUsername, req.GroupName),
		SourceEntityID: ct.Id(req.GroupId),
		SourceService:  "users",
		NeedsAction:    true,
		Payload: map[string]string{
			"inviter_id":   fmt.Sprintf("%d", req.InviterUserId),
			"inviter_name": req.InviterUsername,
			"group_id":     fmt.Sprintf("%d", req.GroupId),
			"group_name":   req.GroupName,
			"action":       "accept_or_decline",
		},
	})

	return notification, nil
}

// CreateGroupInviteForMultipleUsers creates a group invite notification for multiple users
func (s *Server) CreateGroupInviteForMultipleUsers(ctx context.Context, req *pb.CreateGroupInviteForMultipleUsersRequest) (*pb.CreateGroupInviteForMultipleUsersResponse, error) {
	if len(req.InvitedUserIds) == 0 || req.InviterUserId == 0 || req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "invited_user_ids, inviter_user_id, and group_id are required")
	}

	err := s.Application.CreateGroupInviteForMultipleUsers(ctx, req.InvitedUserIds, req.InviterUserId, req.GroupId, req.GroupName, req.InviterUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group invite notifications for multiple users: %v", err)
	}

	// For now, return an empty response since the internal function returns only error
	// In a real implementation, we might want to return the created notifications
	return &pb.CreateGroupInviteForMultipleUsersResponse{}, nil
}

// CreateGroupJoinRequest creates a group join request notification
func (s *Server) CreateGroupJoinRequest(ctx context.Context, req *pb.CreateGroupJoinRequestRequest) (*pb.Notification, error) {
	if req.GroupOwnerId == 0 || req.RequesterUserId == 0 || req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "group_owner_id, requester_user_id, and group_id are required")
	}

	err := s.Application.CreateGroupJoinRequestNotification(ctx, req.GroupOwnerId, req.RequesterUserId, req.GroupId, req.GroupName, req.RequesterUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group join request notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.GroupJoinRequest,
		Title:          "New Group Join Request",
		Message:        fmt.Sprintf("%s wants to join your group \"%s\"", req.RequesterUsername, req.GroupName),
		SourceEntityID: ct.Id(req.GroupId),
		SourceService:  "users",
		NeedsAction:    true,
		Payload: map[string]string{
			"requester_id":   fmt.Sprintf("%d", req.RequesterUserId),
			"requester_name": req.RequesterUsername,
			"group_id":       fmt.Sprintf("%d", req.GroupId),
			"group_name":     req.GroupName,
			"action":         "accept_or_decline",
		},
	})

	return notification, nil
}

// CreateNewEvent creates a new event notification
func (s *Server) CreateNewEvent(ctx context.Context, req *pb.CreateNewEventRequest) (*pb.Notification, error) {
	if req.UserId == 0 || req.GroupId == 0 || req.EventId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, group_id, and event_id are required")
	}

	err := s.Application.CreateNewEventNotification(ctx, req.UserId, req.EventCreatorId, req.GroupId, req.EventId, req.GroupName, req.EventTitle)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create new event notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.NewEvent,
		Title:          "New Event in Group",
		Message:        fmt.Sprintf("New event \"%s\" was created in group \"%s\"", req.EventTitle, req.GroupName),
		SourceEntityID: ct.Id(req.EventId),
		SourceService:  "posts",
		NeedsAction:    false,
		Payload: map[string]string{
			"group_id":    fmt.Sprintf("%d", req.GroupId),
			"group_name":  req.GroupName,
			"event_id":    fmt.Sprintf("%d", req.EventId),
			"event_title": req.EventTitle,
			"action":      "view_event",
		},
	})

	return notification, nil
}

// CreateNewEventForMultipleUsers creates a new event notification for multiple users
func (s *Server) CreateNewEventForMultipleUsers(ctx context.Context, req *pb.CreateNewEventForMultipleUsersRequest) (*pb.CreateNewEventForMultipleUsersResponse, error) {
	if len(req.UserIds) == 0 || req.GroupId == 0 || req.EventId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids, group_id, and event_id are required")
	}

	err := s.Application.CreateNewEventForMultipleUsers(ctx, req.UserIds, req.EventCreatorId, req.GroupId, req.EventId, req.GroupName, req.EventTitle)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create new event notifications for multiple users: %v", err)
	}

	// For now, return an empty response since the internal function returns only error
	// In a real implementation, we might want to return the created notifications
	return &pb.CreateNewEventForMultipleUsersResponse{}, nil
}

// CreatePostLike creates a post like notification
func (s *Server) CreatePostLike(ctx context.Context, req *pb.CreatePostLikeRequest) (*pb.Notification, error) {
	if req.UserId == 0 || req.LikerUserId == 0 || req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, liker_user_id, and post_id are required")
	}

	err := s.Application.CreatePostLikeNotification(ctx, req.UserId, req.LikerUserId, req.PostId, req.LikerUsername, req.Aggregate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create post like notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.PostLike,
		Title:          "Post Liked",
		Message:        fmt.Sprintf("%s liked your post", req.LikerUsername),
		SourceEntityID: ct.Id(req.PostId),
		SourceService:  "posts",
		NeedsAction:    false,
		Payload: map[string]string{
			"liker_id":   fmt.Sprintf("%d", req.LikerUserId),
			"liker_name": req.LikerUsername,
			"post_id":    fmt.Sprintf("%d", req.PostId),
			"action":     "view_post",
		},
	})

	return notification, nil
}

// CreatePostComment creates a post comment notification
func (s *Server) CreatePostComment(ctx context.Context, req *pb.CreatePostCommentRequest) (*pb.Notification, error) {
	if req.UserId == 0 || req.CommenterUserId == 0 || req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, commenter_user_id, and post_id are required")
	}

	err := s.Application.CreatePostCommentNotification(ctx, req.UserId, req.CommenterUserId, req.PostId, req.CommenterUsername, req.CommentContent, req.Aggregate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create post comment notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.PostComment,
		Title:          "New Comment",
		Message:        fmt.Sprintf("%s commented on your post", req.CommenterUsername),
		SourceEntityID: ct.Id(req.PostId),
		SourceService:  "posts",
		NeedsAction:    false,
		Payload: map[string]string{
			"commenter_id":    fmt.Sprintf("%d", req.CommenterUserId),
			"commenter_name":  req.CommenterUsername,
			"post_id":         fmt.Sprintf("%d", req.PostId),
			"comment_content": req.CommentContent,
			"action":          "view_post",
		},
	})

	return notification, nil
}

// CreateMention creates a mention notification
func (s *Server) CreateMention(ctx context.Context, req *pb.CreateMentionRequest) (*pb.Notification, error) {
	if req.UserId == 0 || req.MentionerUserId == 0 || req.PostId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, mentioner_user_id, and post_id are required")
	}

	err := s.Application.CreateMentionNotification(ctx, req.UserId, req.MentionerUserId, req.PostId, req.MentionerUsername, req.PostContent, req.MentionText)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create mention notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.Mention,
		Title:          "You were mentioned",
		Message:        fmt.Sprintf("%s mentioned you in a post", req.MentionerUsername),
		SourceEntityID: ct.Id(req.PostId),
		SourceService:  "posts",
		NeedsAction:    false,
		Payload: map[string]string{
			"mentioner_id":   fmt.Sprintf("%d", req.MentionerUserId),
			"mentioner_name": req.MentionerUsername,
			"post_id":        fmt.Sprintf("%d", req.PostId),
			"post_content":   req.PostContent,
			"mention_text":   req.MentionText,
			"action":         "view_post",
		},
	})

	return notification, nil
}

// CreateNewMessage creates a new message notification
func (s *Server) CreateNewMessage(ctx context.Context, req *pb.CreateNewMessageRequest) (*pb.Notification, error) {
	if req.UserId == 0 || req.SenderUserId == 0 || req.ChatId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_id, sender_user_id, and chat_id are required")
	}

	err := s.Application.CreateNewMessageNotification(ctx, req.UserId, req.SenderUserId, req.ChatId, req.SenderUsername, req.MessageContent, req.Aggregate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create new message notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.NewMessage,
		Title:          "New Message",
		Message:        fmt.Sprintf("%s sent you a message", req.SenderUsername),
		SourceEntityID: ct.Id(req.ChatId),
		SourceService:  "chat",
		NeedsAction:    false,
		Payload: map[string]string{
			"sender_id":       fmt.Sprintf("%d", req.SenderUserId),
			"sender_name":     req.SenderUsername,
			"chat_id":         fmt.Sprintf("%d", req.ChatId),
			"message_content": req.MessageContent,
			"action":          "view_chat",
		},
	})

	return notification, nil
}

// CreateNewMessageForMultipleUsers creates a new message notification for multiple users
func (s *Server) CreateNewMessageForMultipleUsers(ctx context.Context, req *pb.CreateNewMessageForMultipleUsersRequest) (*pb.CreateNewMessageForMultipleUsersResponse, error) {
	if len(req.UserIds) == 0 || req.SenderUserId == 0 || req.ChatId == 0 {
		return nil, status.Error(codes.InvalidArgument, "user_ids, sender_user_id, and chat_id are required")
	}

	err := s.Application.CreateNewMessageForMultipleUsers(ctx, req.UserIds, req.SenderUserId, req.ChatId, req.SenderUsername, req.MessageContent, req.Aggregate)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create new message notifications for multiple users: %v", err)
	}

	// For now, return an empty response since the internal function returns only error
	// In a real implementation, we might want to return the created notifications
	return &pb.CreateNewMessageForMultipleUsersResponse{}, nil
}

// CreateFollowRequestAccepted creates a follow request accepted notification
func (s *Server) CreateFollowRequestAccepted(ctx context.Context, req *pb.CreateFollowRequestAcceptedRequest) (*pb.Notification, error) {
	if req.RequesterUserId == 0 || req.TargetUserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "requester_user_id and target_user_id are required")
	}

	err := s.Application.CreateFollowRequestAcceptedNotification(ctx, req.RequesterUserId, req.TargetUserId, req.TargetUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create follow request accepted notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.FollowRequestAccepted,
		Title:          "Follow Request Accepted",
		Message:        fmt.Sprintf("%s accepted your follow request", req.TargetUsername),
		SourceEntityID: ct.Id(req.TargetUserId),
		SourceService:  "users",
		NeedsAction:    false,
		Payload: map[string]string{
			"target_id":   fmt.Sprintf("%d", req.TargetUserId),
			"target_name": req.TargetUsername,
		},
	})

	return notification, nil
}

// CreateFollowRequestRejected creates a follow request rejected notification
func (s *Server) CreateFollowRequestRejected(ctx context.Context, req *pb.CreateFollowRequestRejectedRequest) (*pb.Notification, error) {
	if req.RequesterUserId == 0 || req.TargetUserId == 0 {
		return nil, status.Error(codes.InvalidArgument, "requester_user_id and target_user_id are required")
	}

	err := s.Application.CreateFollowRequestRejectedNotification(ctx, req.RequesterUserId, req.TargetUserId, req.TargetUsername)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create follow request rejected notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.FollowRequestRejected,
		Title:          "Follow Request Rejected",
		Message:        fmt.Sprintf("%s rejected your follow request", req.TargetUsername),
		SourceEntityID: ct.Id(req.TargetUserId),
		SourceService:  "users",
		NeedsAction:    false,
		Payload: map[string]string{
			"target_id":   fmt.Sprintf("%d", req.TargetUserId),
			"target_name": req.TargetUsername,
		},
	})

	return notification, nil
}

// CreateGroupInviteAccepted creates a group invite accepted notification
func (s *Server) CreateGroupInviteAccepted(ctx context.Context, req *pb.CreateGroupInviteAcceptedRequest) (*pb.Notification, error) {
	if req.InviterUserId == 0 || req.InvitedUserId == 0 || req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "inviter_user_id, invited_user_id, and group_id are required")
	}

	err := s.Application.CreateGroupInviteAcceptedNotification(ctx, req.InviterUserId, req.InvitedUserId, req.GroupId, req.InvitedUsername, req.GroupName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group invite accepted notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.GroupInviteAccepted,
		Title:          "Group Invite Accepted",
		Message:        fmt.Sprintf("%s accepted your group invitation to '%s'", req.InvitedUsername, req.GroupName),
		SourceEntityID: ct.Id(req.GroupId),
		SourceService:  "users",
		NeedsAction:    false,
		Payload: map[string]string{
			"invited_id":   fmt.Sprintf("%d", req.InvitedUserId),
			"invited_name": req.InvitedUsername,
			"group_id":     fmt.Sprintf("%d", req.GroupId),
			"group_name":   req.GroupName,
		},
	})

	return notification, nil
}

// CreateGroupInviteRejected creates a group invite rejected notification
func (s *Server) CreateGroupInviteRejected(ctx context.Context, req *pb.CreateGroupInviteRejectedRequest) (*pb.Notification, error) {
	if req.InviterUserId == 0 || req.InvitedUserId == 0 || req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "inviter_user_id, invited_user_id, and group_id are required")
	}

	err := s.Application.CreateGroupInviteRejectedNotification(ctx, req.InviterUserId, req.InvitedUserId, req.GroupId, req.InvitedUsername, req.GroupName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group invite rejected notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.GroupInviteRejected,
		Title:          "Group Invite Rejected",
		Message:        fmt.Sprintf("%s declined your group invitation to '%s'", req.InvitedUsername, req.GroupName),
		SourceEntityID: ct.Id(req.GroupId),
		SourceService:  "users",
		NeedsAction:    false,
		Payload: map[string]string{
			"invited_id":   fmt.Sprintf("%d", req.InvitedUserId),
			"invited_name": req.InvitedUsername,
			"group_id":     fmt.Sprintf("%d", req.GroupId),
			"group_name":   req.GroupName,
		},
	})

	return notification, nil
}

// CreateGroupJoinRequestAccepted creates a group join request accepted notification
func (s *Server) CreateGroupJoinRequestAccepted(ctx context.Context, req *pb.CreateGroupJoinRequestAcceptedRequest) (*pb.Notification, error) {
	if req.RequesterUserId == 0 || req.GroupOwnerId == 0 || req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "requester_user_id, group_owner_id, and group_id are required")
	}

	err := s.Application.CreateGroupJoinRequestAcceptedNotification(ctx, req.RequesterUserId, req.GroupOwnerId, req.GroupId, req.GroupName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group join request accepted notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.GroupJoinRequestAccepted,
		Title:          "Group Join Request Accepted",
		Message:        fmt.Sprintf("Your request to join group '%s' was approved", req.GroupName),
		SourceEntityID: ct.Id(req.GroupId),
		SourceService:  "users",
		NeedsAction:    false,
		Payload: map[string]string{
			"group_owner_id": fmt.Sprintf("%d", req.GroupOwnerId),
			"group_id":       fmt.Sprintf("%d", req.GroupId),
			"group_name":     req.GroupName,
		},
	})

	return notification, nil
}

// CreateGroupJoinRequestRejected creates a group join request rejected notification
func (s *Server) CreateGroupJoinRequestRejected(ctx context.Context, req *pb.CreateGroupJoinRequestRejectedRequest) (*pb.Notification, error) {
	if req.RequesterUserId == 0 || req.GroupOwnerId == 0 || req.GroupId == 0 {
		return nil, status.Error(codes.InvalidArgument, "requester_user_id, group_owner_id, and group_id are required")
	}

	err := s.Application.CreateGroupJoinRequestRejectedNotification(ctx, req.RequesterUserId, req.GroupOwnerId, req.GroupId, req.GroupName)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create group join request rejected notification: %v", err)
	}

	// Return a basic notification as the internal function returns only error
	notification := s.convertToProtoNotification(&application.Notification{
		Type:           application.GroupJoinRequestRejected,
		Title:          "Group Join Request Rejected",
		Message:        fmt.Sprintf("Your request to join group '%s' was declined", req.GroupName),
		SourceEntityID: ct.Id(req.GroupId),
		SourceService:  "users",
		NeedsAction:    false,
		Payload: map[string]string{
			"group_owner_id": fmt.Sprintf("%d", req.GroupOwnerId),
			"group_id":       fmt.Sprintf("%d", req.GroupId),
			"group_name":     req.GroupName,
		},
	})

	return notification, nil
}

// GetNotificationPreferences returns notification preferences for a user -0 not yet
func (s *Server) GetNotificationPreferences(ctx context.Context, req *wrapperspb.Int64Value) (*pb.NotificationPreferences, error) {
	// For now, return default preferences
	// In a real implementation, this would fetch from a user preferences table
	defaultPrefs := make(map[string]bool)
	for _, notifType := range pb.NotificationType_name {
		defaultPrefs[notifType] = true
	}

	return &pb.NotificationPreferences{
		UserId:      req.Value,
		Preferences: defaultPrefs,
	}, nil
}

// UpdateNotificationPreferences updates notification preferences for a user
func (s *Server) UpdateNotificationPreferences(ctx context.Context, req *pb.UpdateNotificationPreferencesRequest) (*emptypb.Empty, error) {
	// For now, just return success
	// In a real implementation, this would update a user preferences table
	return &emptypb.Empty{}, nil
}

// convertToProtoNotification converts our internal notification model to protobuf format
func (s *Server) convertToProtoNotification(notification *application.Notification) *pb.Notification {
	// Convert map[string]string to map[string]string (which protobuf handles as map<string, string>)
	payload := make(map[string]string)
	for k, v := range notification.Payload {
		payload[k] = v
	}

	var createdAt *timestamppb.Timestamp
	if !notification.CreatedAt.IsZero() {
		createdAt = timestamppb.New(notification.CreatedAt)
	}

	var expiresAt *timestamppb.Timestamp
	if notification.ExpiresAt != nil {
		expiresAt = timestamppb.New(*notification.ExpiresAt)
	}

	var status pb.NotificationStatus
	if notification.Seen {
		status = pb.NotificationStatus_NOTIFICATION_STATUS_READ
	} else {
		status = pb.NotificationStatus_NOTIFICATION_STATUS_UNREAD
	}

	// If deleted_at is not nil, the notification is deleted
	if notification.DeletedAt != nil {
		status = pb.NotificationStatus_NOTIFICATION_STATUS_DELETED
	}

	return &pb.Notification{
		Id:             notification.ID.Int64(),
		UserId:         notification.UserID.Int64(),
		Type:           string(notification.Type),
		Title:          notification.Title,
		Message:        notification.Message,
		SourceService:  notification.SourceService,
		SourceEntityId: notification.SourceEntityID.Int64(),
		NeedsAction:    notification.NeedsAction,
		Acted:          notification.Acted,
		Count:          notification.Count,
		Payload:        payload,
		CreatedAt:      createdAt,
		ExpiresAt:      expiresAt,
		Status:         status,
	}
}

// Helper function to determine if a notification type should be aggregated
func shouldAggregateNotification(notificationType pb.NotificationType) bool {
	switch notificationType {
	case pb.NotificationType_NOTIFICATION_TYPE_POST_LIKE:
		return true
	case pb.NotificationType_NOTIFICATION_TYPE_POST_COMMENT:
		return true
	case pb.NotificationType_NOTIFICATION_TYPE_NEW_FOLLOWER:
		return true
	case pb.NotificationType_NOTIFICATION_TYPE_NEW_MESSAGE:
		return true
	case pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST_ACCEPTED:
		return false // Follow request responses are specific to each request
	case pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST_REJECTED:
		return false // Follow request responses are specific to each request
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE_ACCEPTED:
		return false // Group invite responses are specific to each invitation
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE_REJECTED:
		return false // Group invite responses are specific to each invitation
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST_ACCEPTED:
		return false // Group join request responses are specific to each request
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST_REJECTED:
		return false // Group join request responses are specific to each request
	default:
		return false
	}
}

// convertApplicationNotificationTypeToProto converts application notification type to protobuf notification type
func convertApplicationNotificationTypeToProto(appType application.NotificationType) pb.NotificationType {
	switch appType {
	case application.FollowRequest:
		return pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST
	case application.NewFollower:
		return pb.NotificationType_NOTIFICATION_TYPE_NEW_FOLLOWER
	case application.GroupInvite:
		return pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE
	case application.GroupJoinRequest:
		return pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST
	case application.NewEvent:
		return pb.NotificationType_NOTIFICATION_TYPE_NEW_EVENT
	case application.PostLike:
		return pb.NotificationType_NOTIFICATION_TYPE_POST_LIKE
	case application.PostComment:
		return pb.NotificationType_NOTIFICATION_TYPE_POST_COMMENT
	case application.Mention:
		return pb.NotificationType_NOTIFICATION_TYPE_MENTION
	case application.FollowRequestAccepted:
		return pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST_ACCEPTED
	case application.FollowRequestRejected:
		return pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST_REJECTED
	case application.GroupInviteAccepted:
		return pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE_ACCEPTED
	case application.GroupInviteRejected:
		return pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE_REJECTED
	case application.GroupJoinRequestAccepted:
		return pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST_ACCEPTED
	case application.GroupJoinRequestRejected:
		return pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST_REJECTED
	default:
		return pb.NotificationType_NOTIFICATION_TYPE_UNSPECIFIED
	}
}

// convertProtoNotificationTypeToApplication converts protobuf notification type to application notification type
func convertProtoNotificationTypeToApplication(protoType pb.NotificationType) application.NotificationType {
	switch protoType {
	case pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST:
		return application.FollowRequest
	case pb.NotificationType_NOTIFICATION_TYPE_NEW_FOLLOWER:
		return application.NewFollower
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE:
		return application.GroupInvite
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST:
		return application.GroupJoinRequest
	case pb.NotificationType_NOTIFICATION_TYPE_NEW_EVENT:
		return application.NewEvent
	case pb.NotificationType_NOTIFICATION_TYPE_POST_LIKE:
		return application.PostLike
	case pb.NotificationType_NOTIFICATION_TYPE_POST_COMMENT:
		return application.PostComment
	case pb.NotificationType_NOTIFICATION_TYPE_MENTION:
		return application.Mention
	case pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST_ACCEPTED:
		return application.FollowRequestAccepted
	case pb.NotificationType_NOTIFICATION_TYPE_FOLLOW_REQUEST_REJECTED:
		return application.FollowRequestRejected
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE_ACCEPTED:
		return application.GroupInviteAccepted
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_INVITE_REJECTED:
		return application.GroupInviteRejected
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST_ACCEPTED:
		return application.GroupJoinRequestAccepted
	case pb.NotificationType_NOTIFICATION_TYPE_GROUP_JOIN_REQUEST_REJECTED:
		return application.GroupJoinRequestRejected
	default:
		return application.NotificationType("")
	}
}
