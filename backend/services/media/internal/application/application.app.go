package application

import (
	"context"
	"social-network/services/media/internal/client"
	"social-network/services/media/internal/configs"
	"social-network/services/media/internal/db/dbservice"

	postgresql "social-network/shared/go/postgre"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TxRunner defines the interface for running database transactions
type TxRunner interface {
	RunTx(ctx context.Context, fn func(*dbservice.Queries) error) error
}

// Holds logic for requests and calls
type MediaService struct {
	Pool     *pgxpool.Pool
	S3       S3Service
	Queries  dbservice.Querier
	txRunner TxRunner
	Cfgs     configs.Config
}

func NewMediaService(
	pool *pgxpool.Pool,
	clients *client.Clients,
	queries dbservice.Querier,
	cfgs configs.Config,
) (*MediaService, error) {
	var txRunner TxRunner
	var err error
	if pool != nil {
		queries, ok := queries.(*dbservice.Queries)
		if !ok {
			panic("db must be *dbservice.Queries for transaction support")
		}
		txRunner, err = postgresql.NewPgxTxRunner(pool, queries)
		if err != nil {
			return nil, err
		}
	}
	return &MediaService{
		Pool:     pool,
		S3:       clients,
		Queries:  queries,
		txRunner: txRunner,
		Cfgs:     cfgs,
	}, nil
}
