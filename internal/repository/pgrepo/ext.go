package pgrepo

import (
	"context"
	"github.com/jmoiron/sqlx"
)

type extContext = sqlx.ExtContext

func ext(ctx context.Context, db *sqlx.DB) extContext {
	if tx, ok := txFromCtx(ctx); ok && tx != nil {
		return tx
	}
	return db
}
