package database

import (
	"context"
	"log/slog"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Transaction[T any] struct {
	pool *pgxpool.Pool
}

func NewTransaction[T any](pool *pgxpool.Pool) *Transaction[T] {
	return &Transaction[T]{pool}
}

func (t *Transaction[T]) Execute(ctx context.Context, f func(context.Context) (*T, error)) (*T, error) {
	slog.Info("[Transaction]*****Begin*****")
	tx, err := t.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		slog.Error("[Transaction]fail to begin tran", slog.Any("err", err))
		return nil, err
	}

	defer func() {
		if r := recover(); r != nil {
			slog.Info("[Transaction]--rollback due to runtime panic")
			tx.Rollback(ctx)
			panic(r) // re-throw after rollback
		}
		slog.Info("[Transaction]*****End*****")
	}()

	slog.Info("[Transaction]--executing...")

	result, err := f(ctx)
	if err != nil {
		slog.Info("[Transaction]--rollback to release locked by tran", slog.Any("err", err))
		tx.Rollback(ctx)
		return nil, err
	}

	slog.Info("[Transaction]--commit tran")
	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return result, nil
}
