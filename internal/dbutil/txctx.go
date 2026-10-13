package dbutil

import (
	"context"

	"github.com/jackc/pgx/v5"
)

type txKey struct{}

// WithTx stores tx in ctx so that dbutil.FromCtx picks it up in repositories.
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey{}, tx)
}

// TxFromCtx retrieves the active transaction from ctx, if any.
func TxFromCtx(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey{}).(pgx.Tx)
	return tx, ok
}
