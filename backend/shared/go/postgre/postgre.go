package postgresql

import (
	"context"
	"errors"
	"fmt"
	ce "social-network/shared/go/commonerrors"
	tele "social-network/shared/go/telemetry"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// DBTX is the minimal interface required by sqlc-generated queries.
//
// It is implemented by both *pgxpool.Pool and pgx.Tx, allowing the same
// queries to run inside or outside a transaction.
// type DBTX interface {
// 	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
// 	Query(context.Context, string, ...any) (pgx.Rows, error)
// 	QueryRow(context.Context, string, ...any) pgx.Row
// }

type HasWithTx[T any] interface {
	WithTx(pgx.Tx) T
}

// PgxTxRunner is the production implementation using pgxpool.
type PgxTxRunner[T HasWithTx[T]] struct {
	pool *pgxpool.Pool
	db   T
}

var (
	ErrNilPassed     = errors.New("Passed nil argument")
	ErrNoMoreRetries = errors.New("maximum retries exceeded for serializable transaction")
)

// NewPgxTxRunner creates a new transaction runner.
func NewPgxTxRunner[T HasWithTx[T]](pool *pgxpool.Pool, db T) (*PgxTxRunner[T], error) {
	if pool == nil {
		return nil, ErrNilPassed
	}
	return &PgxTxRunner[T]{
		pool: pool,
		db:   db,
	}, nil
}

// RunTx runs a function inside a database transaction.
func (r *PgxTxRunner[T]) RunTx(ctx context.Context, fn func(T) error) error {
	// start tx.
	tele.Info(ctx, "starting transaction")
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		tele.Error(ctx, "failed to begin transaction @1", "error", err.Error())
		return ce.Wrap(ce.ErrInternal, err, "run tx error")
	}
	defer tx.Rollback(ctx)

	// create queries with transaction - returns *sqlc.Queries.
	qtx := r.db.WithTx(tx)

	// run the function, passing qtx as sqlc.Querier interface.
	if err := fn(qtx); err != nil {
		//TODO add no rows check to avoid logging non error
		tele.Error(ctx, "querier @1", "error", err.Error())
		return err
	}

	// commit transaction.
	tele.Info(ctx, "committing transaction")
	err = tx.Commit(ctx)
	if err != nil {
		return ce.Wrap(ce.ErrInternal, err, "transaction commit error")
	}
	return nil
}

// RunTxSerializable runs a function inside a serializable transaction.
func (r *PgxTxRunner[T]) RunTxSerializable(
	ctx context.Context,
	fn func(T) error,
) error {
	const maxRetries = 3

	for attempt := range maxRetries {
		tele.Info(ctx, "starting serializable transaction", "attempt", attempt+1)

		// start tx with serializable isolation
		tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{
			IsoLevel: pgx.Serializable,
		})
		if err != nil {
			tele.Error(ctx, "failed to begin serializable transaction", "error", err.Error())
			return ce.Wrap(ce.ErrInternal, err, "run serializable tx error")
		}

		// ensure rollback in case of early exit
		defer tx.Rollback(ctx)

		// create sqlc queries tied to transaction
		qtx := r.db.WithTx(tx)

		// run user function
		err = fn(qtx)
		if err != nil {
			// log and return other errors
			tele.Error(ctx, "serializable tx function failed", "error", err.Error())
			return err
		}

		// attempt to commit
		err = tx.Commit(ctx)
		if err != nil {
			// check for serialization failure (Postgres code 40001)
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "40001" {
				tele.Warn(ctx, "serialization failure, retrying transaction @1 @2", "attempt", attempt+1, "error", err.Error())
				// retry loop
				continue
			}
			return ce.Wrap(ce.ErrInternal, err, "serializable transaction commit error")
		}

		// success
		tele.Info(ctx, "serializable transaction committed successfully", "attempt", attempt+1)
		return nil
	}

	return ce.New(ce.ErrInternal, ErrNoMoreRetries)
}

// NewPool creates a pgx connection pool.
//
// Arguments:
//   - ctx: context used to initialize the pool.
//   - address: PostgreSQL connection string.
//
// Returns:
//   - DBTX: pool exposed as a DBTX interface.
//   - *pgxpool.Pool: concrete pool for transaction creation and shutdown.
//   - error: non-nil if the pool cannot be created.
//
// Usage:
//
//	dbtx, pool, err := NewPool(ctx, dsn)
func NewPool(ctx context.Context, address string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, address)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed: %v", err)
	}

	return pool, nil
}
