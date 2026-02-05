package application

import (
	"context"
	"social-network/services/users/internal/client"
	ds "social-network/services/users/internal/db/dbservice"
	"social-network/shared/go/kafgo"
	"social-network/shared/go/models"
	"social-network/shared/go/notifevents"
	"social-network/shared/go/retrievemedia"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TxRunner defines the interface for running database transactions
type TxRunner interface {
	RunTx(ctx context.Context, fn func(*ds.Queries) error) error
}

type Application struct {
	db             ds.Querier
	txRunner       TxRunner
	clients        ClientsInterface
	mediaRetriever *retrievemedia.MediaRetriever
	eventProducer  *notifevents.EventCreator
}

// NewApplication constructs a new UserService
func NewApplication(db ds.Querier, txRunner TxRunner, pool *pgxpool.Pool, clients *client.Clients, eventProducer *kafgo.KafkaProducer) *Application {
	mediaRetriever := retrievemedia.NewMediaRetriever(clients.MediaClient, clients.RedisClient, 3*time.Minute)
	return &Application{
		db:             db,
		txRunner:       txRunner,
		clients:        clients,
		mediaRetriever: mediaRetriever,
		eventProducer:  notifevents.NewEventProducer(eventProducer),
	}
}

// ClientsInterface defines the methods that Application needs from clients.
type ClientsInterface interface {
	GetObj(ctx context.Context, key string, dest any) error
	SetObj(ctx context.Context, key string, value any, exp time.Duration) error
	Del(ctx context.Context, key string) error
	CreateNotification(ctx context.Context, req models.CreateNotificationRequest) error
	CreateFollowRequestNotification(ctx context.Context, targetUserID, requesterUserID int64, requesterUsername string) error
	CreateNewFollower(ctx context.Context, targetUserID, followerUserID int64, followerUsername string) error
	CreateGroupInvite(ctx context.Context, invitedUserId, inviterUserId, groupId int64, groupName, inviterUsername string) error
	CreateGroupJoinRequest(ctx context.Context, groupOnwerId, requesterId, groupId int64, groupName, requesterUsername string) error
	CreateFollowRequestAccepted(ctx context.Context, requesterId, targetUserId int64, targetUsername string) error
	CreateFollowRequestRejected(ctx context.Context, requesterId, targetUserId int64, targetUsername string) error
	CreateGroupInviteAccepted(ctx context.Context, invitedUserId, inviterUserId, groupId int64, groupName, invitedUsername string) error
	CreateGroupInviteRejected(ctx context.Context, invitedUserId, inviterUserId, groupId int64, groupName, invitedUsername string) error
	CreateGroupJoinRequestAccepted(ctx context.Context, requesterId, groupOwnerId, groupId int64, groupName string) error
	CreateGroupJoinRequestRejected(ctx context.Context, requesterId, groupOwnerId, groupId int64, groupName string) error

	// CreateGroupConversation(ctx context.Context, groupId int64, ownerId int64) error
	// CreatePrivateConversation(ctx context.Context, userId1, userId2 int64) error
	// AddMembersToGroupConversation(ctx context.Context, groupId int64, userIds []int64) error
	// DeleteConversationByExactMembers(ctx context.Context, userIds []int64) error
}
