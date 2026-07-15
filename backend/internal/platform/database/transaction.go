package database

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
)

type Tx struct {
	tx pgx.Tx
	*Queries
}

func (t *Tx) Commit(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return nil
	}

	return t.tx.Commit(ctx)
}

func (t *Tx) Rollback(ctx context.Context) error {
	if t == nil || t.tx == nil {
		return nil
	}

	if err := t.tx.Rollback(ctx); err != nil && !errors.Is(err, pgx.ErrTxClosed) {
		return err
	}

	return nil
}

func (s *Store) WithTx(ctx context.Context, fn func(*Tx) error) (err error) {
	tx, err := s.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	err = tx.Commit(ctx)
	return err
}

func (s *Store) WithTxOptions(ctx context.Context, opts pgx.TxOptions, fn func(*Tx) error) (err error) {
	tx, err := s.BeginTx(ctx, opts)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = errors.Join(err, rollbackErr)
			}
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	err = tx.Commit(ctx)
	return err
}
