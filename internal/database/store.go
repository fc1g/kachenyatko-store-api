package database

import (
	"context"

	"github.com/fc1g/kachenyatko-store-api/internal/database/sqlc"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	*sqlc.Queries
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		Queries: sqlc.New(pool),
		pool:    pool,
	}
}

func (s *Store) InTx(ctx context.Context, fn func(q sqlc.Querier) error) error {
	return pgx.BeginFunc(ctx, s.pool, func(tx pgx.Tx) error {
		return fn(s.WithTx(tx))
	})
}
