package events

import (
	"context"
	pb "social-network/shared/gen-go/notifications"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// IntegrationTestSuite runs integration tests for the event handler
type IntegrationTestSuite struct {
	suite.Suite
	eventHandler *EventHandler
	mockApp      *MockApplication
}

// SetupTest runs before each test
func (suite *IntegrationTestSuite) SetupTest() {
	suite.mockApp = new(MockApplication)
	suite.eventHandler = &EventHandler{App: suite.mockApp}
}

// Test that all event types can be processed successfully
func (suite *IntegrationTestSuite) TestProcessEventTypesIntegration() {
	ctx := context.Background()

	// Test PostCommentCreated
	suite.mockApp.On("CreatePostCommentNotification", mock.Anything, int64(123), int64(456), int64(123), "user1", "comment body", true).Return(nil)
	event1 := &pb.NotificationEvent{
		EventId:   "test-1",
		EventType: pb.EventType_POST_COMMENT_CREATED,
		Payload: &pb.NotificationEvent_PostCommentCreated{
			PostCommentCreated: &pb.PostCommentCreated{
				PostCreatorId:     123,
				CommenterUserId:   456,
				PostId:            123,
				CommenterUsername: "user1",
				Body:              "comment body",
				Aggregate:         true,
			},
		},
	}
	err := suite.eventHandler.Handle(ctx, event1)
	assert.NoError(suite.T(), err)

	// Test PostLiked
	suite.mockApp.On("CreatePostLikeNotification", mock.Anything, int64(123), int64(456), int64(123), "user2", true).Return(nil)
	event2 := &pb.NotificationEvent{
		EventId:   "test-2",
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
	err = suite.eventHandler.Handle(ctx, event2)
	assert.NoError(suite.T(), err)

	// Test FollowRequestCreated
	suite.mockApp.On("CreateFollowRequestNotification", mock.Anything, int64(123), int64(456), "user3").Return(nil)
	event3 := &pb.NotificationEvent{
		EventId:   "test-3",
		EventType: pb.EventType_FOLLOW_REQUEST_CREATED,
		Payload: &pb.NotificationEvent_FollowRequestCreated{
			FollowRequestCreated: &pb.FollowRequestCreated{
				TargetUserId:      123,
				RequesterUserId:   456,
				RequesterUsername: "user3",
			},
		},
	}
	err = suite.eventHandler.Handle(ctx, event3)
	assert.NoError(suite.T(), err)

	// Test FollowRequestCancelled
	suite.mockApp.On("DeleteFollowRequestNotification", mock.Anything, int64(123), int64(456)).Return(nil)
	event4 := &pb.NotificationEvent{
		EventId:   "test-4",
		EventType: pb.EventType_FOLLOW_REQUEST_CANCELLED,
		Payload: &pb.NotificationEvent_FollowRequestCancelled{
			FollowRequestCancelled: &pb.FollowRequestCancelled{
				TargetUserId:    123,
				RequesterUserId: 456,
			},
		},
	}
	err = suite.eventHandler.Handle(ctx, event4)
	assert.NoError(suite.T(), err)

	// Test GroupJoinRequestCancelled
	suite.mockApp.On("DeleteGroupJoinRequestNotification", mock.Anything, int64(123), int64(456), int64(789)).Return(nil)
	event5 := &pb.NotificationEvent{
		EventId:   "test-5",
		EventType: pb.EventType_GROUP_JOIN_REQUEST_CANCELLED,
		Payload: &pb.NotificationEvent_GroupJoinRequestCancelled{
			GroupJoinRequestCancelled: &pb.GroupJoinRequestCancelled{
				GroupOwnerId:    123,
				RequesterUserId: 456,
				GroupId:         789,
			},
		},
	}
	err = suite.eventHandler.Handle(ctx, event5)
	assert.NoError(suite.T(), err)

	// Test NewFollowerCreated
	suite.mockApp.On("CreateNewFollowerNotification", mock.Anything, int64(123), int64(456), "user4", true).Return(nil)
	event6 := &pb.NotificationEvent{
		EventId:   "test-6",
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
	err = suite.eventHandler.Handle(ctx, event6)
	assert.NoError(suite.T(), err)

	// Test GroupInviteCreated
	suite.mockApp.On("CreateGroupInviteForMultipleUsers", mock.Anything, []int64{123}, int64(456), int64(789), "group1", "user5").Return(nil)
	event7 := &pb.NotificationEvent{
		EventId:   "test-7",
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
	err = suite.eventHandler.Handle(ctx, event7)
	assert.NoError(suite.T(), err)

	// Verify all expectations were met
	suite.mockApp.AssertExpectations(suite.T())
}

// Test error handling when application layer returns an error
func (suite *IntegrationTestSuite) TestErrorHandlingIntegration() {
	ctx := context.Background()

	// Set up expectation to return an error
	expectedErr := assert.AnError
	suite.mockApp.On("CreatePostCommentNotification", mock.Anything, int64(123), int64(456), int64(123), "user1", "comment body", true).Return(expectedErr)

	event := &pb.NotificationEvent{
		EventId:   "error-test",
		EventType: pb.EventType_POST_COMMENT_CREATED,
		Payload: &pb.NotificationEvent_PostCommentCreated{
			PostCommentCreated: &pb.PostCommentCreated{
				PostCreatorId:     123,
				CommenterUserId:   456,
				PostId:            123,
				CommenterUsername: "user1",
				Body:              "comment body",
				Aggregate:         true,
			},
		},
	}

	err := suite.eventHandler.Handle(ctx, event)
	assert.Equal(suite.T(), expectedErr, err)
	suite.mockApp.AssertExpectations(suite.T())
}

// Test processing multiple events in sequence
func (suite *IntegrationTestSuite) TestSequentialProcessingIntegration() {
	ctx := context.Background()

	events := []*pb.NotificationEvent{
		{
			EventId:   "seq-1",
			EventType: pb.EventType_POST_COMMENT_CREATED,
			Payload: &pb.NotificationEvent_PostCommentCreated{
				PostCommentCreated: &pb.PostCommentCreated{
					PostCreatorId:     123,
					CommenterUserId:   456,
					PostId:            123,
					CommenterUsername: "user1",
					Body:              "comment body",
					Aggregate:         true,
				},
			},
		},
		{
			EventId:   "seq-2",
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
		},
		{
			EventId:   "seq-3",
			EventType: pb.EventType_FOLLOW_REQUEST_CREATED,
			Payload: &pb.NotificationEvent_FollowRequestCreated{
				FollowRequestCreated: &pb.FollowRequestCreated{
					TargetUserId:      123,
					RequesterUserId:   456,
					RequesterUsername: "user3",
				},
			},
		},
		{
			EventId:   "seq-4",
			EventType: pb.EventType_FOLLOW_REQUEST_CANCELLED,
			Payload: &pb.NotificationEvent_FollowRequestCancelled{
				FollowRequestCancelled: &pb.FollowRequestCancelled{
					TargetUserId:    123,
					RequesterUserId: 456,
				},
			},
		},
	}

	// Set up expectations for all events
	suite.mockApp.On("CreatePostCommentNotification", mock.Anything, int64(123), int64(456), int64(123), "user1", "comment body", true).Return(nil)
	suite.mockApp.On("CreatePostLikeNotification", mock.Anything, int64(123), int64(456), int64(123), "user2", true).Return(nil)
	suite.mockApp.On("CreateFollowRequestNotification", mock.Anything, int64(123), int64(456), "user3").Return(nil)
	suite.mockApp.On("DeleteFollowRequestNotification", mock.Anything, int64(123), int64(456)).Return(nil)

	// Process all events sequentially
	for _, event := range events {
		err := suite.eventHandler.Handle(ctx, event)
		assert.NoError(suite.T(), err)
	}

	suite.mockApp.AssertExpectations(suite.T())
}

// Test that the event handler can handle cancellation events specifically
func (suite *IntegrationTestSuite) TestCancellationEventsIntegration() {
	ctx := context.Background()

	// Test Follow Request Cancellation
	suite.mockApp.On("DeleteFollowRequestNotification", mock.Anything, int64(1001), int64(2001)).Return(nil)
	followCancelEvent := &pb.NotificationEvent{
		EventId:   "cancel-follow-1",
		EventType: pb.EventType_FOLLOW_REQUEST_CANCELLED,
		Payload: &pb.NotificationEvent_FollowRequestCancelled{
			FollowRequestCancelled: &pb.FollowRequestCancelled{
				TargetUserId:    1001,
				RequesterUserId: 2001,
			},
		},
	}
	err := suite.eventHandler.Handle(ctx, followCancelEvent)
	assert.NoError(suite.T(), err)

	// Test Group Join Request Cancellation
	suite.mockApp.On("DeleteGroupJoinRequestNotification", mock.Anything, int64(1001), int64(2001), int64(3001)).Return(nil)
	groupCancelEvent := &pb.NotificationEvent{
		EventId:   "cancel-group-join-1",
		EventType: pb.EventType_GROUP_JOIN_REQUEST_CANCELLED,
		Payload: &pb.NotificationEvent_GroupJoinRequestCancelled{
			GroupJoinRequestCancelled: &pb.GroupJoinRequestCancelled{
				GroupOwnerId:    1001,
				RequesterUserId: 2001,
				GroupId:         3001,
			},
		},
	}
	err = suite.eventHandler.Handle(ctx, groupCancelEvent)
	assert.NoError(suite.T(), err)

	suite.mockApp.AssertExpectations(suite.T())
}

// TearDownTest runs after each test
func (suite *IntegrationTestSuite) TearDownTest() {
	suite.mockApp.AssertExpectations(suite.T())
}

// Run the suite
func TestIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
