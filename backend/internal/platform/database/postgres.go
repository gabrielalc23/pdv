package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNilPool = errors.New("database pool is nil")

type Store struct {
	Pool    *pgxpool.Pool
	Queries *Queries
}

func Open(ctx context.Context, dsn string) (*Store, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return NewStore(pool), nil
}

func NewStore(pool *pgxpool.Pool) *Store {
	if pool == nil {
		return &Store{}
	}

	return &Store{
		Pool:    pool,
		Queries: New(pool),
	}
}

func (s *Store) Close() {
	if s != nil && s.Pool != nil {
		s.Pool.Close()
	}
}

func (s *Store) BeginTx(ctx context.Context, opts pgx.TxOptions) (*Tx, error) {
	if s == nil || s.Pool == nil {
		return nil, ErrNilPool
	}

	tx, err := s.Pool.BeginTx(ctx, opts)
	if err != nil {
		return nil, err
	}

	return &Tx{
		tx:      tx,
		Queries: s.Queries.WithTx(tx),
	}, nil
}
