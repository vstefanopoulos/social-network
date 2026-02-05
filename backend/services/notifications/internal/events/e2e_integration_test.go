package events

import (
	"context"
	pb "social-network/shared/gen-go/notifications"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// EndToEndIntegrationTestSuite tests the complete flow from event reception to processing
type EndToEndIntegrationTestSuite struct {
	suite.Suite
	eventHandler *EventHandler
	mockApp      *MockApplication
	ctx          context.Context
}

// SetupTest runs before each test
func (suite *EndToEndIntegrationTestSuite) SetupTest() {
	suite.ctx = context.Background()
	suite.mockApp = new(MockApplication)
	suite.eventHandler = &EventHandler{App: suite.mockApp}
}

// TestCompleteEventFlow tests the complete flow of receiving and processing events
func (suite *EndToEndIntegrationTestSuite) TestCompleteEventFlow() {
	// Simulate a realistic sequence of events that might happen in the system
	events := []*pb.NotificationEvent{
		// User A follows User B
		{
			EventId:   "event-1",
			EventType: pb.EventType_FOLLOW_REQUEST_CREATED,
			Metadata:  map[string]string{"source": "users-service"},
			Payload: &pb.NotificationEvent_FollowRequestCreated{
				FollowRequestCreated: &pb.FollowRequestCreated{
					TargetUserId:      1001,
					RequesterUserId:   2001,
					RequesterUsername: "user_a",
				},
			},
		},
		// User A cancels the follow request
		{
			EventId:   "event-2",
			EventType: pb.EventType_FOLLOW_REQUEST_CANCELLED,
			Metadata:  map[string]string{"source": "users-service"},
			Payload: &pb.NotificationEvent_FollowRequestCancelled{
				FollowRequestCancelled: &pb.FollowRequestCancelled{
					TargetUserId:    1001,
					RequesterUserId: 2001,
				},
			},
		},
		// User B posts something
		{
			EventId:   "event-3",
			EventType: pb.EventType_NEW_EVENT_CREATED,
			Metadata:  map[string]string{"source": "posts-service"},
			Payload: &pb.NotificationEvent_NewEventCreated{
				NewEventCreated: &pb.NewEventCreated{
					UserId:         []int64{1001},
					EventCreatorId: 1001,
					GroupId:        3001,
					EventId:        4001,
					GroupName:      "Tech Enthusiasts",
					EventTitle:     "Weekly Tech Meetup",
				},
			},
		},
		// Someone comments on User B's post
		{
			EventId:   "event-4",
			EventType: pb.EventType_POST_COMMENT_CREATED,
			Metadata:  map[string]string{"source": "posts-service"},
			Payload: &pb.NotificationEvent_PostCommentCreated{
				PostCommentCreated: &pb.PostCommentCreated{
					PostCreatorId:     1001,
					PostId:            5001,
					CommentId:         6001,
					CommenterUserId:   3001,
					CommenterUsername: "user_c",
					Body:              "Great post! Thanks for sharing.",
					Aggregate:         true,
				},
			},
		},
		// Someone likes User B's post
		{
			EventId:   "event-5",
			EventType: pb.EventType_POST_LIKED,
			Metadata:  map[string]string{"source": "posts-service"},
			Payload: &pb.NotificationEvent_PostLiked{
				PostLiked: &pb.PostLiked{
					EntityCreatorId: 1001,
					PostId:          5001,
					LikerUserId:     4001,
					LikerUsername:   "user_d",
					Aggregate:       true,
				},
			},
		},
		// User B gets a message
		{
			EventId:   "event-6",
			EventType: pb.EventType_NEW_MESSAGE_CREATED,
			Metadata:  map[string]string{"source": "chat-service"},
			Payload: &pb.NotificationEvent_NewMessageCreated{
				NewMessageCreated: &pb.NewMessageCreated{
					UserId:         []int64{1001},
					SenderUserId:   5001,
					ChatId:         7001,
					SenderUsername: "user_e",
					MessageContent: "Hey, are we still meeting tomorrow?",
					Aggregate:      true,
				},
			},
		},
		// Group join request
		{
			EventId:   "event-7",
			EventType: pb.EventType_GROUP_JOIN_REQUEST_CREATED,
			Metadata:  map[string]string{"source": "users-service"},
			Payload: &pb.NotificationEvent_GroupJoinRequestCreated{
				GroupJoinRequestCreated: &pb.GroupJoinRequestCreated{
					GroupOwnerId:      1001,
					RequesterUserId:   2001,
					GroupId:           3001,
					GroupName:         "Tech Enthusiasts",
					RequesterUsername: "user_a",
				},
			},
		},
		// User A cancels the group join request
		{
			EventId:   "event-8",
			EventType: pb.EventType_GROUP_JOIN_REQUEST_CANCELLED,
			Metadata:  map[string]string{"source": "users-service"},
			Payload: &pb.NotificationEvent_GroupJoinRequestCancelled{
				GroupJoinRequestCancelled: &pb.GroupJoinRequestCancelled{
					GroupOwnerId:    1001,
					RequesterUserId: 2001,
					GroupId:         3001,
				},
			},
		},
	}

	// Set up expectations for all events
	suite.mockApp.On("CreateFollowRequestNotification",
		mock.Anything, int64(1001), int64(2001), "user_a").Return(nil)
	suite.mockApp.On("DeleteFollowRequestNotification",
		mock.Anything, int64(1001), int64(2001)).Return(nil)
	suite.mockApp.On("CreateNewEventForMultipleUsers",
		mock.Anything, []int64{1001}, int64(1001), int64(3001), int64(4001), "Tech Enthusiasts", "Weekly Tech Meetup").Return(nil)
	suite.mockApp.On("CreatePostCommentNotification",
		mock.Anything, int64(1001), int64(3001), int64(5001), "user_c", "Great post! Thanks for sharing.", true).Return(nil)
	suite.mockApp.On("CreatePostLikeNotification",
		mock.Anything, int64(1001), int64(4001), int64(5001), "user_d", true).Return(nil)
	suite.mockApp.On("CreateNewMessageForMultipleUsers",
		mock.Anything, []int64{1001}, int64(5001), int64(7001), "user_e", "Hey, are we still meeting tomorrow?", true).Return(nil)
	suite.mockApp.On("CreateGroupJoinRequestNotification",
		mock.Anything, int64(1001), int64(2001), int64(3001), "Tech Enthusiasts", "user_a").Return(nil)
	suite.mockApp.On("DeleteGroupJoinRequestNotification",
		mock.Anything, int64(1001), int64(2001), int64(3001)).Return(nil)

	// Process all events in sequence (simulating real Kafka consumption)
	for i, event := range events {
		err := suite.eventHandler.Handle(suite.ctx, event)
		assert.NoError(suite.T(), err, "Event %d should be processed successfully", i+1)
	}

	// Verify all expectations were met
	suite.mockApp.AssertExpectations(suite.T())
}

// TestRealWorldScenario tests a realistic scenario with mixed event types
func (suite *EndToEndIntegrationTestSuite) TestRealWorldScenario() {
	// Simulate a busy day on the social network
	scenarioEvents := []*pb.NotificationEvent{
		// Morning: Several likes on yesterday's post
		{
			EventId:   "morning-1",
			EventType: pb.EventType_POST_LIKED,
			Payload: &pb.NotificationEvent_PostLiked{
				PostLiked: &pb.PostLiked{
					EntityCreatorId: 1001,
					PostId:          1001,
					LikerUserId:     2001,
					LikerUsername:   "alice",
					Aggregate:       true,
				},
			},
		},
		{
			EventId:   "morning-2",
			EventType: pb.EventType_POST_LIKED,
			Payload: &pb.NotificationEvent_PostLiked{
				PostLiked: &pb.PostLiked{
					EntityCreatorId: 1001,
					PostId:          1001,
					LikerUserId:     2002,
					LikerUsername:   "bob",
					Aggregate:       true,
				},
			},
		},
		// Mid-day: New followers
		{
			EventId:   "midday-1",
			EventType: pb.EventType_NEW_FOLLOWER_CREATED,
			Payload: &pb.NotificationEvent_NewFollowerCreated{
				NewFollowerCreated: &pb.NewFollowerCreated{
					TargetUserId:     1001,
					FollowerUserId:   3001,
					FollowerUsername: "charlie",
					Aggregate:        true,
				},
			},
		},
		// Afternoon: Group activity
		{
			EventId:   "afternoon-1",
			EventType: pb.EventType_NEW_EVENT_CREATED,
			Payload: &pb.NotificationEvent_NewEventCreated{
				NewEventCreated: &pb.NewEventCreated{
					UserId:         []int64{1001},
					EventCreatorId: 1001,
					GroupId:        4001,
					EventId:        5001,
					GroupName:      "Photography Club",
					EventTitle:     "Sunset Photography Session",
				},
			},
		},
		// Evening: Chat messages
		{
			EventId:   "evening-1",
			EventType: pb.EventType_NEW_MESSAGE_CREATED,
			Payload: &pb.NotificationEvent_NewMessageCreated{
				NewMessageCreated: &pb.NewMessageCreated{
					UserId:         []int64{1001},
					SenderUserId:   6001,
					ChatId:         7001,
					SenderUsername: "diana",
					MessageContent: "Did you see the new camera gear?",
					Aggregate:      true,
				},
			},
		},
		// Late evening: Follow request cancellation
		{
			EventId:   "late-evening-1",
			EventType: pb.EventType_FOLLOW_REQUEST_CANCELLED,
			Payload: &pb.NotificationEvent_FollowRequestCancelled{
				FollowRequestCancelled: &pb.FollowRequestCancelled{
					TargetUserId:    1001,
					RequesterUserId: 8001,
				},
			},
		},
	}

	// Set up expectations for the scenario
	suite.mockApp.On("CreatePostLikeNotification", mock.Anything, int64(1001), int64(2001), int64(1001), "alice", true).Return(nil)
	suite.mockApp.On("CreatePostLikeNotification", mock.Anything, int64(1001), int64(2002), int64(1001), "bob", true).Return(nil)
	suite.mockApp.On("CreateNewFollowerNotification", mock.Anything, int64(1001), int64(3001), "charlie", true).Return(nil)
	suite.mockApp.On("CreateNewEventForMultipleUsers", mock.Anything, []int64{1001}, int64(1001), int64(4001), int64(5001), "Photography Club", "Sunset Photography Session").Return(nil)
	suite.mockApp.On("CreateNewMessageForMultipleUsers", mock.Anything, []int64{1001}, int64(6001), int64(7001), "diana", "Did you see the new camera gear?", true).Return(nil)
	suite.mockApp.On("DeleteFollowRequestNotification", mock.Anything, int64(1001), int64(8001)).Return(nil)

	// Process all events in sequence
	for i, event := range scenarioEvents {
		err := suite.eventHandler.Handle(suite.ctx, event)
		assert.NoError(suite.T(), err, "Scenario event %d should be processed successfully", i+1)
	}

	// Verify expectations
	suite.mockApp.AssertExpectations(suite.T())
}

// TestCancellationSpecificScenario focuses on cancellation events
func (suite *EndToEndIntegrationTestSuite) TestCancellationSpecificScenario() {
	cancellationEvents := []*pb.NotificationEvent{
		// Follow request is made
		{
			EventId:   "follow-request-1",
			EventType: pb.EventType_FOLLOW_REQUEST_CREATED,
			Payload: &pb.NotificationEvent_FollowRequestCreated{
				FollowRequestCreated: &pb.FollowRequestCreated{
					TargetUserId:      1001,
					RequesterUserId:   2001,
					RequesterUsername: "requester_user",
				},
			},
		},
		// Follow request is cancelled
		{
			EventId:   "follow-cancel-1",
			EventType: pb.EventType_FOLLOW_REQUEST_CANCELLED,
			Payload: &pb.NotificationEvent_FollowRequestCancelled{
				FollowRequestCancelled: &pb.FollowRequestCancelled{
					TargetUserId:    1001,
					RequesterUserId: 2001,
				},
			},
		},
		// Group join request is made
		{
			EventId:   "group-join-request-1",
			EventType: pb.EventType_GROUP_JOIN_REQUEST_CREATED,
			Payload: &pb.NotificationEvent_GroupJoinRequestCreated{
				GroupJoinRequestCreated: &pb.GroupJoinRequestCreated{
					GroupOwnerId:      1001,
					RequesterUserId:   2001,
					GroupId:           3001,
					GroupName:         "Test Group",
					RequesterUsername: "requester_user",
				},
			},
		},
		// Group join request is cancelled
		{
			EventId:   "group-join-cancel-1",
			EventType: pb.EventType_GROUP_JOIN_REQUEST_CANCELLED,
			Payload: &pb.NotificationEvent_GroupJoinRequestCancelled{
				GroupJoinRequestCancelled: &pb.GroupJoinRequestCancelled{
					GroupOwnerId:    1001,
					RequesterUserId: 2001,
					GroupId:         3001,
				},
			},
		},
	}

	// Set up expectations for cancellation scenario
	suite.mockApp.On("CreateFollowRequestNotification", mock.Anything, int64(1001), int64(2001), "requester_user").Return(nil)
	suite.mockApp.On("DeleteFollowRequestNotification", mock.Anything, int64(1001), int64(2001)).Return(nil)
	suite.mockApp.On("CreateGroupJoinRequestNotification", mock.Anything, int64(1001), int64(2001), int64(3001), "Test Group", "requester_user").Return(nil)
	suite.mockApp.On("DeleteGroupJoinRequestNotification", mock.Anything, int64(1001), int64(2001), int64(3001)).Return(nil)

	// Process cancellation events
	for i, event := range cancellationEvents {
		err := suite.eventHandler.Handle(suite.ctx, event)
		assert.NoError(suite.T(), err, "Cancellation event %d should be processed successfully", i+1)
	}

	// Verify expectations
	suite.mockApp.AssertExpectations(suite.T())
}

// TearDownTest runs after each test
func (suite *EndToEndIntegrationTestSuite) TearDownTest() {
	suite.mockApp.AssertExpectations(suite.T())
}

// Run the suite
func TestEndToEndIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(EndToEndIntegrationTestSuite))
}
