package application

import (
	"context"
	"fmt"
	"social-network/services/chat/internal/client"
	"social-network/services/chat/internal/db/dbservice"
	ce "social-network/shared/go/commonerrors"
	ct "social-network/shared/go/ct"
	postgresql "social-network/shared/go/postgre"
	"social-network/shared/go/retrieveusers"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

// TxRunner defines the interface for running database transactions
type TxRunner interface {
	RunTx(ctx context.Context, fn func(*dbservice.Queries) error) error
	RunTxSerializable(
		ctx context.Context,
		fn func(*dbservice.Queries) error,
	) error
}

// Holds logic for requests and calls
type ChatService struct {
	Clients      Clients
	RetriveUsers *retrieveusers.UserRetriever
	Queries      dbservice.Querier
	txRunner     TxRunner
	NatsConn     *nats.Conn
}

type Clients interface {
	// Verifies that user with userId is a member of group with groupId.
	IsGroupMember(ctx context.Context,
		groupId ct.Id, userId ct.Id) (bool, *ce.Error)

	// Returns true if either user is following the other.
	AreConnected(ctx context.Context, userA, userB ct.Id) (bool, *ce.Error)
}

func NewChatService(
	pool *pgxpool.Pool,
	clients *client.Clients,
	queries dbservice.Querier,
	userRetriever *retrieveusers.UserRetriever,
	natsConn *nats.Conn,
) (*ChatService, error) {
	var txRunner TxRunner
	var err error
	if pool != nil {
		queries, ok := queries.(*dbservice.Queries)
		if !ok {
			panic("db must be *db.Queries for transaction support")
		}
		txRunner, err = postgresql.NewPgxTxRunner(pool, queries)
		if err != nil {
			return nil, fmt.Errorf("failed to create pgxTxRunner %w", err)
		}
	}

	return &ChatService{
		Clients:      clients,
		Queries:      queries,
		txRunner:     txRunner,
		RetriveUsers: userRetriever,
		NatsConn:     natsConn,
	}, nil
}
