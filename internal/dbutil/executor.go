package dbutil

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Executor is satisfied by both *pgxpool.Pool and pgx.Tx, allowing repositories
// to run queries identically whether or not a transaction is active.
type Executor interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// FromCtx returns the transaction stored in ctx if one is present; otherwise
// it returns the fallback (typically the connection pool).
func FromCtx(ctx context.Context, fallback Executor) Executor {
	if tx, ok := ctx.Value(txKey{}).(pgx.Tx); ok {
		return tx
	}
	return fallback
}
