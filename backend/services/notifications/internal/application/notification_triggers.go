package application

import (
	"context"
	"fmt"
	tele "social-network/shared/go/telemetry"
)

// CreateFollowRequestNotification creates a notification when a user sends a follow request to a private account
func (a *Application) CreateFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64, requesterUsername string) error {
	title := "New Follow Request"
	message := fmt.Sprintf("%s wants to follow you", requesterUsername)

	payload := map[string]string{
		"requester_id":   fmt.Sprintf("%d", requesterUserID),
		"requester_name": requesterUsername,
	}

	// Follow requests should not be aggregated since they require action
	// We ignore the aggregate parameter for notifications that need action
	_, err := a.CreateNotificationWithAggregation(
		ctx,
		targetUserID,    // recipient
		FollowRequest,   // type
		title,           // title
		message,         // message
		"users",         // source service
		requesterUserID, // source entity ID
		true,            // needs action
		payload,         // payload
		false,           // never aggregate follow requests
	)
	if err != nil {
		return fmt.Errorf("failed to create follow request notification: %w", err)
	}

	return nil
}

// CreateNewFollowerNotification creates a notification when someone follows a user
func (a *Application) CreateNewFollowerNotification(ctx context.Context, targetUserID, followerUserID int64, followerUsername string, aggregate bool) error {
	title := "New Follower"
	message := fmt.Sprintf("%s is now following you", followerUsername)

	payload := map[string]string{
		"follower_id":   fmt.Sprintf("%d", followerUserID),
		"follower_name": followerUsername,
	}

	_, err := a.CreateNotificationWithAggregation(
		ctx,
		targetUserID,   // recipient
		NewFollower,    // type
		title,          // title
		message,        // message
		"users",        // source service
		followerUserID, // source entity ID
		false,          // doesn't need action
		payload,        // payload
		aggregate,      // whether to aggregate
	)
	if err != nil {
		return fmt.Errorf("failed to create new follower notification: %w", err)
	}

	return nil
}

// CreateGroupInviteNotification creates a notification when a user is invited to join a group
func (a *Application) CreateGroupInviteNotification(ctx context.Context, invitedUserID, inviterUserID, groupID int64, groupName, inviterUsername string) error {
	return a.CreateGroupInviteForMultipleUsers(ctx, []int64{invitedUserID}, inviterUserID, groupID, groupName, inviterUsername)
}

// CreateGroupInviteForMultipleUsers creates a notification when users are invited to join a group
func (a *Application) CreateGroupInviteForMultipleUsers(ctx context.Context, invitedUserIDs []int64, inviterUserID, groupID int64, groupName, inviterUsername string) error {
	title := "Group Invitation"
	message := fmt.Sprintf("%s invited you to join the group \"%s\"", inviterUsername, groupName)

	payload := map[string]string{
		"inviter_id":   fmt.Sprintf("%d", inviterUserID),
		"inviter_name": inviterUsername,
		"group_id":     fmt.Sprintf("%d", groupID),
		"group_name":   groupName,
		"action":       "accept_or_decline",
	}

	// Create notifications for each user
	for _, invitedUserID := range invitedUserIDs {
		_, err := a.CreateNotificationWithAggregation(
			ctx,
			invitedUserID, // recipient
			GroupInvite,   // type
			title,         // title
			message,       // message
			"users",       // source service
			groupID,       // source entity ID (the group)
			true,          // needs action
			payload,       // payload
			false,         // never aggregate group invites
		)
		if err != nil {
			return fmt.Errorf("failed to create group invite notification for user %d: %w", invitedUserID, err)
		}
	}

	return nil
}

// CreateGroupJoinRequestNotification creates a notification when someone requests to join a group
func (a *Application) CreateGroupJoinRequestNotification(ctx context.Context, groupOwnerID, requesterID, groupID int64, groupName, requesterUsername string) error {
	title := "New Group Join Request"
	message := fmt.Sprintf("%s wants to join your group \"%s\"", requesterUsername, groupName)

	payload := map[string]string{
		"requester_id":   fmt.Sprintf("%d", requesterID),
		"requester_name": requesterUsername,
		"group_id":       fmt.Sprintf("%d", groupID),
		"group_name":     groupName,
		"action":         "accept_or_decline",
	}

	_, err := a.CreateNotificationWithAggregation(
		ctx,
		groupOwnerID,     // recipient (group owner)
		GroupJoinRequest, // type
		title,            // title
		message,          // message
		"users",          // source service
		groupID,          // source entity ID (the group)
		true,             // needs action
		payload,          // payload
		false,            // don't aggregate group join requests
	)
	if err != nil {
		return fmt.Errorf("failed to create group join request notification: %w", err)
	}

	return nil
}

// CreateNewEventNotification creates a notification when a new event is created in a group the user is part of
func (a *Application) CreateNewEventNotification(ctx context.Context, userID, eventCreatorID, groupID, eventID int64, groupName, eventTitle string) error {
	tele.Debug(ctx, "create new event notification called")
	return a.CreateNewEventForMultipleUsers(ctx, []int64{userID}, eventCreatorID, groupID, eventID, groupName, eventTitle)
}

// CreateNewEventForMultipleUsers creates a notification when a new event is created in a group for multiple users
func (a *Application) CreateNewEventForMultipleUsers(ctx context.Context, userIDs []int64, eventCreatorID, groupID, eventID int64, groupName, eventTitle string) error {
	title := fmt.Sprintf("New Event: %s", eventTitle)
	message := fmt.Sprintf("New event \"%s\" was created in group \"%s\"", eventTitle, groupName)

	payload := map[string]string{
		"group_id":    fmt.Sprintf("%d", groupID),
		"group_name":  groupName,
		"event_id":    fmt.Sprintf("%d", eventID),
		"event_title": eventTitle,
		"action":      "view_event",
	}

	// Create notifications for each user
	for _, userID := range userIDs {
		if userID == eventCreatorID {
			continue
		}
		_, err := a.CreateNotificationWithAggregation(
			ctx,
			userID,   // recipient
			NewEvent, // type
			title,    // title
			message,  // message
			"posts",  // source service
			eventID,  // source entity ID (the event)
			false,    // doesn't need action (just informational)
			payload,  // payload
			false,    // never aggregate new events
		)
		if err != nil {
			return fmt.Errorf("failed to create new event notification for user %d: %w", userID, err)
		}
	}

	return nil
}

// Additional notification types for extended functionality

// CreatePostLikeNotification creates a notification when someone likes a user's post
func (a *Application) CreatePostLikeNotification(ctx context.Context, userID, likerID, postID int64, likerUsername string, aggregate bool) error {
	title := "Post Liked"
	message := fmt.Sprintf("%s liked your post", likerUsername)

	payload := map[string]string{
		"liker_id":   fmt.Sprintf("%d", likerID),
		"liker_name": likerUsername,
		"post_id":    fmt.Sprintf("%d", postID),
		"action":     "view_post",
	}

	_, err := a.CreateNotificationWithAggregation(
		ctx,
		userID,    // recipient
		PostLike,  // type
		title,     // title
		message,   // message
		"posts",   // source service
		postID,    // source entity ID
		false,     // doesn't need action
		payload,   // payload
		aggregate, // whether to aggregate
	)
	if err != nil {
		return fmt.Errorf("failed to create post like notification: %w", err)
	}

	return nil
}

// CreatePostCommentNotification creates a notification when someone comments on a user's post
func (a *Application) CreatePostCommentNotification(ctx context.Context, userID, commenterID, postID int64, commenterUsername, postContent string, aggregate bool) error {
	title := "New Comment"
	message := fmt.Sprintf("%s commented on your post", commenterUsername)

	payload := map[string]string{
		"commenter_id":   fmt.Sprintf("%d", commenterID),
		"commenter_name": commenterUsername,
		"post_id":        fmt.Sprintf("%d", postID),
		"post_content":   postContent,
		"action":         "view_post",
	}

	_, err := a.CreateNotificationWithAggregation(
		ctx,
		userID,      // recipient
		PostComment, // type
		title,       // title
		message,     // message
		"posts",     // source service
		postID,      // source entity ID
		false,       // doesn't need action
		payload,     // payload
		aggregate,   // whether to aggregate
	)
	if err != nil {
		return fmt.Errorf("failed to create post comment notification: %w", err)
	}

	return nil
}

// CreateMentionNotification creates a notification when a user is mentioned in a post or comment
func (a *Application) CreateMentionNotification(ctx context.Context, userID, mentionerID, postID int64, mentionerUsername, postContent, mentionText string) error {
	title := "You were mentioned"
	message := fmt.Sprintf("%s mentioned you in a post", mentionerUsername)

	payload := map[string]string{
		"mentioner_id":   fmt.Sprintf("%d", mentionerID),
		"mentioner_name": mentionerUsername,
		"post_id":        fmt.Sprintf("%d", postID),
		"post_content":   postContent,
		"mention_text":   mentionText,
		"action":         "view_post",
	}

	_, err := a.CreateNotificationWithAggregation(
		ctx,
		userID,  // recipient
		Mention, // type
		title,   // title
		message, // message
		"posts", // source service
		postID,  // source entity ID
		false,   // doesn't need action
		payload, // payload
		false,   // never aggregate mentions
	)
	if err != nil {
		return fmt.Errorf("failed to create mention notification: %w", err)
	}

	return nil
}

// CreateNewMessageNotification creates a notification when a user receives a new message in a chat
func (a *Application) CreateNewMessageNotification(ctx context.Context, userID, senderID, chatID int64, senderUsername, messageContent string, aggregate bool) error {
	return a.CreateNewMessageForMultipleUsers(ctx, []int64{userID}, senderID, chatID, senderUsername, messageContent, aggregate)
}

// CreateNewMessageForMultipleUsers creates a notification when a user sends a message to multiple users in a chat
func (a *Application) CreateNewMessageForMultipleUsers(ctx context.Context, userIDs []int64, senderID, chatID int64, senderUsername, messageContent string, aggregate bool) error {
	title := "New Message"
	message := fmt.Sprintf("%s sent you a message", senderUsername)

	payload := map[string]string{
		"sender_id":       fmt.Sprintf("%d", senderID),
		"sender_name":     senderUsername,
		"chat_id":         fmt.Sprintf("%d", chatID),
		"message_content": messageContent,
		"action":          "view_chat",
	}

	// Create notifications for each user
	for _, userID := range userIDs {
		_, err := a.CreateNotificationWithAggregation(
			ctx,
			userID,     // recipient
			NewMessage, // type
			title,      // title
			message,    // message
			"chat",     // source service
			chatID,     // source entity ID (the chat)
			false,      // doesn't need action (just informational)
			payload,    // payload
			aggregate,  // whether to aggregate
		)
		if err != nil {
			return fmt.Errorf("failed to create new message notification for user %d: %w", userID, err)
		}
	}

	return nil
}

// CreateFollowRequestAcceptedNotification creates a notification when a follow request is accepted
func (a *Application) CreateFollowRequestAcceptedNotification(ctx context.Context, requesterUserID, targetUserID int64, targetUsername string) error {
	title := "Follow Request Accepted"
	message := fmt.Sprintf("%s accepted your follow request", targetUsername)

	payload := map[string]string{
		"target_id":    fmt.Sprintf("%d", targetUserID),
		"target_name":  targetUsername,
		"requester_id": fmt.Sprintf("%d", requesterUserID),
	}

	_, err := a.CreateNotification(
		ctx,
		requesterUserID,       // recipient (the original requester)
		FollowRequestAccepted, // type
		title,                 // title
		message,               // message
		"users",               // source service
		targetUserID,          // source entity ID
		false,                 // doesn't need action (just informational)
		payload,               // payload
	)
	if err != nil {
		return fmt.Errorf("failed to create follow request accepted notification: %w", err)
	}

	return nil
}

// CreateFollowRequestRejectedNotification creates a notification when a follow request is rejected
func (a *Application) CreateFollowRequestRejectedNotification(ctx context.Context, requesterUserID, targetUserID int64, targetUsername string) error {
	title := "Follow Request Rejected"
	message := fmt.Sprintf("%s rejected your follow request", targetUsername)

	payload := map[string]string{
		"target_id":    fmt.Sprintf("%d", targetUserID),
		"target_name":  targetUsername,
		"requester_id": fmt.Sprintf("%d", requesterUserID),
	}

	_, err := a.CreateNotification(
		ctx,
		requesterUserID,       // recipient (the original requester)
		FollowRequestRejected, // type
		title,                 // title
		message,               // message
		"users",               // source service
		targetUserID,          // source entity ID
		false,                 // doesn't need action (just informational)
		payload,               // payload
	)
	if err != nil {
		return fmt.Errorf("failed to create follow request rejected notification: %w", err)
	}

	return nil
}

// CreateGroupInviteAcceptedNotification creates a notification when a group invite is accepted
func (a *Application) CreateGroupInviteAcceptedNotification(ctx context.Context, inviterUserID, invitedUserID, groupID int64, invitedUsername, groupName string) error {
	title := "Group Invite Accepted"
	message := fmt.Sprintf("%s accepted your group invitation to '%s'", invitedUsername, groupName)

	payload := map[string]string{
		"invited_id":   fmt.Sprintf("%d", invitedUserID),
		"invited_name": invitedUsername,
		"group_id":     fmt.Sprintf("%d", groupID),
		"group_name":   groupName,
		"inviter_id":   fmt.Sprintf("%d", inviterUserID),
	}

	_, err := a.CreateNotification(
		ctx,
		inviterUserID,       // recipient (the original inviter)
		GroupInviteAccepted, // type
		title,               // title
		message,             // message
		"users",             // source service
		groupID,             // source entity ID (the group)
		false,               // doesn't need action (just informational)
		payload,             // payload
	)
	if err != nil {
		return fmt.Errorf("failed to create group invite accepted notification: %w", err)
	}

	return nil
}

// CreateGroupInviteRejectedNotification creates a notification when a group invite is rejected
func (a *Application) CreateGroupInviteRejectedNotification(ctx context.Context, inviterUserID, invitedUserID, groupID int64, invitedUsername, groupName string) error {
	title := "Group Invite Rejected"
	message := fmt.Sprintf("%s declined your group invitation to '%s'", invitedUsername, groupName)

	payload := map[string]string{
		"invited_id":   fmt.Sprintf("%d", invitedUserID),
		"invited_name": invitedUsername,
		"group_id":     fmt.Sprintf("%d", groupID),
		"group_name":   groupName,
		"inviter_id":   fmt.Sprintf("%d", inviterUserID),
	}

	_, err := a.CreateNotification(
		ctx,
		inviterUserID,       // recipient (the original inviter)
		GroupInviteRejected, // type
		title,               // title
		message,             // message
		"users",             // source service
		groupID,             // source entity ID (the group)
		false,               // doesn't need action (just informational)
		payload,             // payload
	)
	if err != nil {
		return fmt.Errorf("failed to create group invite rejected notification: %w", err)
	}

	return nil
}

// CreateGroupJoinRequestAcceptedNotification creates a notification when a group join request is accepted
func (a *Application) CreateGroupJoinRequestAcceptedNotification(ctx context.Context, requesterUserID, groupOwnerID, groupID int64, groupName string) error {
	title := "Group Join Request Accepted"
	message := fmt.Sprintf("Your request to join group '%s' was approved", groupName)

	payload := map[string]string{
		"group_owner_id":                   fmt.Sprintf("%d", groupOwnerID),
		"group_id":                         fmt.Sprintf("%d", groupID),
		"group_name":                       groupName,
		"requester_id":                     fmt.Sprintf("%d", requesterUserID),
		"group_owner_notification_user_id": fmt.Sprintf("%d", groupOwnerID),
	}

	_, err := a.CreateNotification(
		ctx,
		requesterUserID,          // recipient (the original requester)
		GroupJoinRequestAccepted, // type
		title,                    // title
		message,                  // message
		"users",                  // source service
		groupID,                  // source entity ID (the group)
		false,                    // doesn't need action (just informational)
		payload,                  // payload
	)
	if err != nil {
		return fmt.Errorf("failed to create group join request accepted notification: %w", err)
	}

	return nil
}

// CreateGroupJoinRequestRejectedNotification creates a notification when a group join request is rejected
func (a *Application) CreateGroupJoinRequestRejectedNotification(ctx context.Context, requesterUserID, groupOwnerID, groupID int64, groupName string) error {
	title := "Group Join Request Rejected"
	message := fmt.Sprintf("Your request to join group '%s' was declined", groupName)

	payload := map[string]string{
		"group_owner_id":                   fmt.Sprintf("%d", groupOwnerID),
		"group_id":                         fmt.Sprintf("%d", groupID),
		"group_name":                       groupName,
		"requester_id":                     fmt.Sprintf("%d", requesterUserID),
		"group_owner_notification_user_id": fmt.Sprintf("%d", groupOwnerID),
	}

	_, err := a.CreateNotification(
		ctx,
		requesterUserID,          // recipient (the original requester)
		GroupJoinRequestRejected, // type
		title,                    // title
		message,                  // message
		"users",                  // source service
		groupID,                  // source entity ID (the group)
		false,                    // doesn't need action (just informational)
		payload,                  // payload
	)
	if err != nil {
		return fmt.Errorf("failed to create group join request rejected notification: %w", err)
	}

	return nil
}
