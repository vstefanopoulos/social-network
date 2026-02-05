package events

import (
	"context"
	"fmt"
	pb "social-network/shared/gen-go/notifications"
	tele "social-network/shared/go/telemetry"
)

// EventHandler handles different types of notification events from Kafka
type EventHandler struct {
	App ApplicationService
}

// Ensure EventHandler implements EventProcessor interface
var _ EventProcessor = (*EventHandler)(nil)

// NewEventHandler creates a new event handler
func NewEventHandler(app ApplicationService) *EventHandler {
	return &EventHandler{
		App: app,
	}
}

// Handle processes a notification event based on its type
func (h *EventHandler) Handle(ctx context.Context, event *pb.NotificationEvent) error {
	switch payload := event.Payload.(type) {
	case *pb.NotificationEvent_PostCommentCreated:
		return h.handlePostCommentCreated(ctx, payload.PostCommentCreated)
	case *pb.NotificationEvent_PostLiked:
		return h.handlePostLiked(ctx, payload.PostLiked)
	case *pb.NotificationEvent_FollowRequestCreated:
		return h.handleFollowRequestCreated(ctx, payload.FollowRequestCreated)
	case *pb.NotificationEvent_NewFollowerCreated:
		return h.handleNewFollowerCreated(ctx, payload.NewFollowerCreated)
	case *pb.NotificationEvent_GroupInviteCreated:
		return h.handleGroupInviteCreated(ctx, payload.GroupInviteCreated)
	case *pb.NotificationEvent_GroupJoinRequestCreated:
		return h.handleGroupJoinRequestCreated(ctx, payload.GroupJoinRequestCreated)
	case *pb.NotificationEvent_NewEventCreated:
		return h.handleNewEventCreated(ctx, payload.NewEventCreated)
	case *pb.NotificationEvent_MentionCreated:
		return h.handleMentionCreated(ctx, payload.MentionCreated)
	case *pb.NotificationEvent_NewMessageCreated:
		return h.handleNewMessageCreated(ctx, payload.NewMessageCreated)
	case *pb.NotificationEvent_FollowRequestAccepted:
		return h.handleFollowRequestAccepted(ctx, payload.FollowRequestAccepted)
	case *pb.NotificationEvent_FollowRequestRejected:
		return h.handleFollowRequestRejected(ctx, payload.FollowRequestRejected)
	case *pb.NotificationEvent_GroupInviteAccepted:
		return h.handleGroupInviteAccepted(ctx, payload.GroupInviteAccepted)
	case *pb.NotificationEvent_GroupInviteRejected:
		return h.handleGroupInviteRejected(ctx, payload.GroupInviteRejected)
	case *pb.NotificationEvent_GroupJoinRequestAccepted:
		return h.handleGroupJoinRequestAccepted(ctx, payload.GroupJoinRequestAccepted)
	case *pb.NotificationEvent_GroupJoinRequestRejected:
		return h.handleGroupJoinRequestRejected(ctx, payload.GroupJoinRequestRejected)
	case *pb.NotificationEvent_FollowRequestCancelled:
		return h.handleFollowRequestCancelled(ctx, payload.FollowRequestCancelled)
	case *pb.NotificationEvent_GroupJoinRequestCancelled:
		return h.handleGroupJoinRequestCancelled(ctx, payload.GroupJoinRequestCancelled)
	default:
		return fmt.Errorf("unknown notification event payload type: %T", payload)
	}
}

// Individual handler methods
func (h *EventHandler) handlePostCommentCreated(ctx context.Context, event *pb.PostCommentCreated) error {
	return h.App.CreatePostCommentNotification(
		ctx,
		event.PostCreatorId,     // userId (post owner)
		event.CommenterUserId,   // commenterId
		event.PostId,            // postId
		event.CommenterUsername, // commenterUsername
		event.Body,              // postContent
		event.Aggregate,         // aggregate - use value from event
	)
}

func (h *EventHandler) handlePostLiked(ctx context.Context, event *pb.PostLiked) error {
	return h.App.CreatePostLikeNotification(
		ctx,
		event.EntityCreatorId, // userId (post owner)
		event.LikerUserId,     // likerId
		event.PostId,          // postId
		event.LikerUsername,   // likerUsername
		event.Aggregate,       // aggregate - use value from event
	)
}

func (h *EventHandler) handleFollowRequestCreated(ctx context.Context, event *pb.FollowRequestCreated) error {
	return h.App.CreateFollowRequestNotification(
		ctx,
		event.TargetUserId,      // targetUserID
		event.RequesterUserId,   // requesterUserID
		event.RequesterUsername, // requesterUsername
	)
}

func (h *EventHandler) handleFollowRequestCancelled(ctx context.Context, event *pb.FollowRequestCancelled) error {
	return h.App.DeleteFollowRequestNotification(
		ctx,
		event.TargetUserId,    // targetUserID
		event.RequesterUserId, // requesterUserID
	)
}

func (h *EventHandler) handleGroupJoinRequestCancelled(ctx context.Context, event *pb.GroupJoinRequestCancelled) error {
	return h.App.DeleteGroupJoinRequestNotification(
		ctx,
		event.GroupOwnerId,    // groupOwnerID
		event.RequesterUserId, // requesterUserID
		event.GroupId,         // groupID
	)
}

func (h *EventHandler) handleNewFollowerCreated(ctx context.Context, event *pb.NewFollowerCreated) error {
	return h.App.CreateNewFollowerNotification(
		ctx,
		event.TargetUserId,     // targetUserID
		event.FollowerUserId,   // followerUserID
		event.FollowerUsername, // followerUsername
		event.Aggregate,        // aggregate - use value from event
	)
}

func (h *EventHandler) handleGroupInviteCreated(ctx context.Context, event *pb.GroupInviteCreated) error {
	return h.App.CreateGroupInviteForMultipleUsers(
		ctx,
		event.InvitedUserId,   // invitedUserID
		event.InviterUserId,   // inviterUserID
		event.GroupId,         // groupID
		event.GroupName,       // groupName
		event.InviterUsername, // inviterUsername
	)
}

func (h *EventHandler) handleGroupJoinRequestCreated(ctx context.Context, event *pb.GroupJoinRequestCreated) error {
	return h.App.CreateGroupJoinRequestNotification(
		ctx,
		event.GroupOwnerId,      // groupOwnerID
		event.RequesterUserId,   // requesterID
		event.GroupId,           // groupID
		event.GroupName,         // groupName
		event.RequesterUsername, // requesterUsername
	)
}

func (h *EventHandler) handleNewEventCreated(ctx context.Context, event *pb.NewEventCreated) error {
	tele.Info(ctx, "handle new event created called with params @1", "params", event)
	return h.App.CreateNewEventForMultipleUsers(
		ctx,
		event.UserId,         // userID
		event.EventCreatorId, //eventCreatorID
		event.GroupId,        // groupID
		event.EventId,        // eventID
		event.GroupName,      // groupName
		event.EventTitle,     // eventTitle
	)
}

func (h *EventHandler) handleMentionCreated(ctx context.Context, event *pb.MentionCreated) error {
	return h.App.CreateMentionNotification(
		ctx,
		event.MentionedUserId,   // userID
		event.MentionerUserId,   // mentionerID
		event.PostId,            // postID
		event.MentionerUsername, // mentionerUsername
		event.PostContent,       // postContent
		event.MentionText,       // mentionText
	)
}

func (h *EventHandler) handleNewMessageCreated(ctx context.Context, event *pb.NewMessageCreated) error {
	return h.App.CreateNewMessageForMultipleUsers(
		ctx,
		event.UserId,         // userID
		event.SenderUserId,   // senderID
		event.ChatId,         // chatID
		event.SenderUsername, // senderUsername
		event.MessageContent, // messageContent
		event.Aggregate,      // aggregate - use value from event
	)
}

func (h *EventHandler) handleFollowRequestAccepted(ctx context.Context, event *pb.FollowRequestAccepted) error {
	return h.App.CreateFollowRequestAcceptedNotification(
		ctx,
		event.RequesterUserId, // requesterUserID
		event.TargetUserId,    // targetUserID
		event.TargetUsername,  // targetUsername
	)
}

func (h *EventHandler) handleFollowRequestRejected(ctx context.Context, event *pb.FollowRequestRejected) error {
	return h.App.CreateFollowRequestRejectedNotification(
		ctx,
		event.RequesterUserId, // requesterUserID
		event.TargetUserId,    // targetUserID
		event.TargetUsername,  // targetUsername
	)
}

func (h *EventHandler) handleGroupInviteAccepted(ctx context.Context, event *pb.GroupInviteAccepted) error {
	return h.App.CreateGroupInviteAcceptedNotification(
		ctx,
		event.InviterUserId,   // inviterUserID
		event.InvitedUserId,   // invitedUserID
		event.GroupId,         // groupID
		event.InvitedUsername, // invitedUsername
		event.GroupName,       // groupName
	)
}

func (h *EventHandler) handleGroupInviteRejected(ctx context.Context, event *pb.GroupInviteRejected) error {
	return h.App.CreateGroupInviteRejectedNotification(
		ctx,
		event.InviterUserId,   // inviterUserID
		event.InvitedUserId,   // invitedUserID
		event.GroupId,         // groupID
		event.InvitedUsername, // invitedUsername
		event.GroupName,       // groupName
	)
}

func (h *EventHandler) handleGroupJoinRequestAccepted(ctx context.Context, event *pb.GroupJoinRequestAccepted) error {
	return h.App.CreateGroupJoinRequestAcceptedNotification(
		ctx,
		event.RequesterUserId, // requesterUserID
		event.GroupOwnerId,    // groupOwnerID
		event.GroupId,         // groupID
		event.GroupName,       // groupName
	)
}

func (h *EventHandler) handleGroupJoinRequestRejected(ctx context.Context, event *pb.GroupJoinRequestRejected) error {
	return h.App.CreateGroupJoinRequestRejectedNotification(
		ctx,
		event.RequesterUserId, // requesterUserID
		event.GroupOwnerId,    // groupOwnerID
		event.GroupId,         // groupID
		event.GroupName,       // groupName
	)
}
