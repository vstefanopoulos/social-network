package events

import (
	"context"
	"social-network/services/notifications/internal/application"
	pb "social-network/shared/gen-go/notifications"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockApplication is a mock that implements the Application interface
type MockApplication struct {
	mock.Mock
}

// Implement all the methods needed for testing
func (m *MockApplication) CreatePostCommentNotification(ctx context.Context, userID, commenterID, postID int64, commenterUsername, commentContent string, aggregate bool) error {
	args := m.Called(ctx, userID, commenterID, postID, commenterUsername, commentContent, aggregate)
	return args.Error(0)
}

func (m *MockApplication) CreatePostLikeNotification(ctx context.Context, userID, likerID, postID int64, likerUsername string, aggregate bool) error {
	args := m.Called(ctx, userID, likerID, postID, likerUsername, aggregate)
	return args.Error(0)
}

func (m *MockApplication) CreateFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64, requesterUsername string) error {
	args := m.Called(ctx, targetUserID, requesterUserID, requesterUsername)
	return args.Error(0)
}

func (m *MockApplication) CreateNewFollowerNotification(ctx context.Context, targetUserID, followerUserID int64, followerUsername string, aggregate bool) error {
	args := m.Called(ctx, targetUserID, followerUserID, followerUsername, aggregate)
	return args.Error(0)
}

func (m *MockApplication) CreateGroupInviteNotification(ctx context.Context, invitedUserID, inviterUserID, groupID int64, groupName, inviterUsername string) error {
	args := m.Called(ctx, invitedUserID, inviterUserID, groupID, groupName, inviterUsername)
	return args.Error(0)
}

func (m *MockApplication) CreateGroupInviteForMultipleUsers(ctx context.Context, invitedUserIDs []int64, inviterUserID, groupID int64, groupName, inviterUsername string) error {
	args := m.Called(ctx, invitedUserIDs, inviterUserID, groupID, groupName, inviterUsername)
	return args.Error(0)
}

func (m *MockApplication) CreateGroupJoinRequestNotification(ctx context.Context, groupOwnerID, requesterID, groupID int64, groupName, requesterUsername string) error {
	args := m.Called(ctx, groupOwnerID, requesterID, groupID, groupName, requesterUsername)
	return args.Error(0)
}

func (m *MockApplication) CreateNewEventNotification(ctx context.Context, userID, eventCreatorID, groupID, eventID int64, groupName, eventTitle string) error {
	args := m.Called(ctx, userID, eventCreatorID, groupID, eventID, groupName, eventTitle)
	return args.Error(0)
}

func (m *MockApplication) CreateNewEventForMultipleUsers(ctx context.Context, userIDs []int64, eventCreatorID int64, groupID, eventID int64, groupName, eventTitle string) error {
	args := m.Called(ctx, userIDs, eventCreatorID, groupID, eventID, groupName, eventTitle)
	return args.Error(0)
}

func (m *MockApplication) CreateMentionNotification(ctx context.Context, userID, mentionerID, postID int64, mentionerUsername, postContent, mentionText string) error {
	args := m.Called(ctx, userID, mentionerID, postID, mentionerUsername, postContent, mentionText)
	return args.Error(0)
}

func (m *MockApplication) CreateNewMessageNotification(ctx context.Context, userID, senderID, chatID int64, senderUsername, messageContent string, aggregate bool) error {
	args := m.Called(ctx, userID, senderID, chatID, senderUsername, messageContent, aggregate)
	return args.Error(0)
}

func (m *MockApplication) CreateNewMessageForMultipleUsers(ctx context.Context, userIDs []int64, senderID, chatID int64, senderUsername, messageContent string, aggregate bool) error {
	args := m.Called(ctx, userIDs, senderID, chatID, senderUsername, messageContent, aggregate)
	return args.Error(0)
}

func (m *MockApplication) CreateFollowRequestAcceptedNotification(ctx context.Context, requesterUserID, targetUserID int64, targetUsername string) error {
	args := m.Called(ctx, requesterUserID, targetUserID, targetUsername)
	return args.Error(0)
}

func (m *MockApplication) CreateFollowRequestRejectedNotification(ctx context.Context, requesterUserID, targetUserID int64, targetUsername string) error {
	args := m.Called(ctx, requesterUserID, targetUserID, targetUsername)
	return args.Error(0)
}

func (m *MockApplication) CreateGroupInviteAcceptedNotification(ctx context.Context, inviterUserID, invitedUserID, groupID int64, invitedUsername, groupName string) error {
	args := m.Called(ctx, inviterUserID, invitedUserID, groupID, invitedUsername, groupName)
	return args.Error(0)
}

func (m *MockApplication) CreateGroupInviteRejectedNotification(ctx context.Context, inviterUserID, invitedUserID, groupID int64, invitedUsername, groupName string) error {
	args := m.Called(ctx, inviterUserID, invitedUserID, groupID, invitedUsername, groupName)
	return args.Error(0)
}

func (m *MockApplication) CreateGroupJoinRequestAcceptedNotification(ctx context.Context, requesterUserID, groupOwnerID, groupID int64, groupName string) error {
	args := m.Called(ctx, requesterUserID, groupOwnerID, groupID, groupName)
	return args.Error(0)
}

func (m *MockApplication) CreateGroupJoinRequestRejectedNotification(ctx context.Context, requesterUserID, groupOwnerID, groupID int64, groupName string) error {
	args := m.Called(ctx, requesterUserID, groupOwnerID, groupID, groupName)
	return args.Error(0)
}

func (m *MockApplication) CreateDefaultNotificationTypes(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockApplication) GetNotification(ctx context.Context, notificationID, userID int64) (*application.Notification, error) {
	args := m.Called(ctx, notificationID, userID)
	return args.Get(0).(*application.Notification), args.Error(1)
}

func (m *MockApplication) GetUserNotifications(ctx context.Context, userID int64, limit, offset int32) ([]*application.Notification, error) {
	args := m.Called(ctx, userID, limit, offset)
	return args.Get(0).([]*application.Notification), args.Error(1)
}

func (m *MockApplication) GetUserNotificationsCount(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockApplication) GetUserUnreadNotificationsCount(ctx context.Context, userID int64) (int64, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockApplication) MarkNotificationAsRead(ctx context.Context, notificationID, userID int64) error {
	args := m.Called(ctx, notificationID, userID)
	return args.Error(0)
}

func (m *MockApplication) MarkAllAsRead(ctx context.Context, userID int64) error {
	args := m.Called(ctx, userID)
	return args.Error(0)
}

func (m *MockApplication) DeleteNotification(ctx context.Context, notificationID, userID int64) error {
	args := m.Called(ctx, notificationID, userID)
	return args.Error(0)
}

func (m *MockApplication) CreateNotification(ctx context.Context, userID int64, notifType application.NotificationType, title, message, sourceService string, sourceEntityID int64, needsAction bool, payload map[string]string) (*application.Notification, error) {
	args := m.Called(ctx, userID, notifType, title, message, sourceService, sourceEntityID, needsAction, payload)
	return args.Get(0).(*application.Notification), args.Error(1)
}

func (m *MockApplication) CreateNotificationWithAggregation(ctx context.Context, userID int64, notifType application.NotificationType, title, message, sourceService string, sourceEntityID int64, needsAction bool, payload map[string]string, aggregate bool) (*application.Notification, error) {
	args := m.Called(ctx, userID, notifType, title, message, sourceService, sourceEntityID, needsAction, payload, aggregate)
	return args.Get(0).(*application.Notification), args.Error(1)
}

func (m *MockApplication) CreateNotifications(ctx context.Context, notifications []struct {
	UserID         int64
	Type           application.NotificationType
	Title          string
	Message        string
	SourceService  string
	SourceEntityID int64
	NeedsAction    bool
	Payload        map[string]string
}) ([]*application.Notification, error) {
	args := m.Called(ctx, notifications)
	return args.Get(0).([]*application.Notification), args.Error(1)
}

func (m *MockApplication) DeleteFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64) error {
	args := m.Called(ctx, targetUserID, requesterUserID)
	return args.Error(0)
}

func (m *MockApplication) DeleteGroupJoinRequestNotification(ctx context.Context, groupOwnerID, requesterUserID, groupID int64) error {
	args := m.Called(ctx, groupOwnerID, requesterUserID, groupID)
	return args.Error(0)
}

// Unit tests for each event handler
func TestEventHandler_HandlePostCommentCreated(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-event-id",
		EventType: pb.EventType_POST_COMMENT_CREATED,
		Payload: &pb.NotificationEvent_PostCommentCreated{
			PostCommentCreated: &pb.PostCommentCreated{
				PostCreatorId:     123,
				CommenterUserId:   789,
				PostId:            123,
				CommenterUsername: "test_user",
				Body:              "This is a test comment",
				Aggregate:         true,
			},
		},
	}

	// Set up expectations
	mockApp.On("CreatePostCommentNotification",
		mock.Anything,
		int64(123),               // userID (post owner)
		int64(789),               // commenterID
		int64(123),               // postID
		"test_user",              // commenterUsername
		"This is a test comment", // commentContent
		true,                     // aggregate
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandlePostLiked(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-event-id",
		EventType: pb.EventType_POST_LIKED,
		Payload: &pb.NotificationEvent_PostLiked{
			PostLiked: &pb.PostLiked{
				EntityCreatorId: 123,
				LikerUserId:     456,
				PostId:          123,
				LikerUsername:   "user2",
				Aggregate:       true,
			},
		},
	}

	// Set up expectations
	mockApp.On("CreatePostLikeNotification",
		mock.Anything,
		int64(123), // userID (post owner)
		int64(456), // likerID
		int64(123), // postID
		"user2",    // likerUsername
		true,       // aggregate
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandleFollowRequestCreated(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-event-id",
		EventType: pb.EventType_FOLLOW_REQUEST_CREATED,
		Payload: &pb.NotificationEvent_FollowRequestCreated{
			FollowRequestCreated: &pb.FollowRequestCreated{
				TargetUserId:      123,
				RequesterUserId:   456,
				RequesterUsername: "user3",
			},
		},
	}

	// Set up expectations
	mockApp.On("CreateFollowRequestNotification",
		mock.Anything,
		int64(123), // targetUserID
		int64(456), // requesterUserID
		"user3",    // requesterUsername
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandleFollowRequestCancelled(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-cancel-event-id",
		EventType: pb.EventType_FOLLOW_REQUEST_CANCELLED,
		Payload: &pb.NotificationEvent_FollowRequestCancelled{
			FollowRequestCancelled: &pb.FollowRequestCancelled{
				TargetUserId:    123,
				RequesterUserId: 456,
			},
		},
	}

	// Set up expectations
	mockApp.On("DeleteFollowRequestNotification",
		mock.Anything,
		int64(123), // targetUserID
		int64(456), // requesterUserID
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandleGroupJoinRequestCancelled(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-cancel-event-id",
		EventType: pb.EventType_GROUP_JOIN_REQUEST_CANCELLED,
		Payload: &pb.NotificationEvent_GroupJoinRequestCancelled{
			GroupJoinRequestCancelled: &pb.GroupJoinRequestCancelled{
				GroupOwnerId:    123,
				RequesterUserId: 456,
				GroupId:         789,
			},
		},
	}

	// Set up expectations
	mockApp.On("DeleteGroupJoinRequestNotification",
		mock.Anything,
		int64(123), // groupOwnerID
		int64(456), // requesterUserID
		int64(789), // groupID
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandleNewFollowerCreated(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-event-id",
		EventType: pb.EventType_NEW_FOLLOWER_CREATED,
		Payload: &pb.NotificationEvent_NewFollowerCreated{
			NewFollowerCreated: &pb.NewFollowerCreated{
				TargetUserId:     123,
				FollowerUserId:   456,
				FollowerUsername: "user4",
				Aggregate:        true,
			},
		},
	}

	// Set up expectations
	mockApp.On("CreateNewFollowerNotification",
		mock.Anything,
		int64(123), // targetUserID
		int64(456), // followerUserID
		"user4",    // followerUsername
		true,       // aggregate
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandleGroupInviteCreated(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-event-id",
		EventType: pb.EventType_GROUP_INVITE_CREATED,
		Payload: &pb.NotificationEvent_GroupInviteCreated{
			GroupInviteCreated: &pb.GroupInviteCreated{
				InvitedUserId:   []int64{123},
				InviterUserId:   456,
				GroupId:         789,
				GroupName:       "group1",
				InviterUsername: "user5",
			},
		},
	}

	// Set up expectations
	mockApp.On("CreateGroupInviteForMultipleUsers",
		mock.Anything,
		[]int64{123}, // invitedUserIDs
		int64(456),   // inviterUserID
		int64(789),   // groupID
		"group1",     // groupName
		"user5",      // inviterUsername
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandleGroupJoinRequestCreated(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-event-id",
		EventType: pb.EventType_GROUP_JOIN_REQUEST_CREATED,
		Payload: &pb.NotificationEvent_GroupJoinRequestCreated{
			GroupJoinRequestCreated: &pb.GroupJoinRequestCreated{
				GroupOwnerId:      123,
				RequesterUserId:   456,
				GroupId:           789,
				GroupName:         "group1",
				RequesterUsername: "user6",
			},
		},
	}

	// Set up expectations
	mockApp.On("CreateGroupJoinRequestNotification",
		mock.Anything,
		int64(123), // groupOwnerID
		int64(456), // requesterID
		int64(789), // groupID
		"group1",   // groupName
		"user6",    // requesterUsername
	).Return(nil)

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.NoError(t, err)
	mockApp.AssertExpectations(t)
}

func TestEventHandler_HandleUnknownEvent(t *testing.T) {
	mockApp := new(MockApplication)
	eventHandler := &EventHandler{App: mockApp}

	event := &pb.NotificationEvent{
		EventId:   "test-event-id",
		EventType: pb.EventType_EVENT_TYPE_UNSPECIFIED,
		Payload:   nil, // No payload
	}

	// Execute
	err := eventHandler.Handle(context.Background(), event)

	// Assert
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown notification event payload type")
}
