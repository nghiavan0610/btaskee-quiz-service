package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// QueryExecutor provides methods to execute queries without prepared statement caching
// This helps avoid "cached plan must not change result type" errors for dynamic queries
type QueryExecutor struct {
	pool *pgxpool.Pool
}

// NewQueryExecutor creates a new query executor
func NewQueryExecutor(pool *pgxpool.Pool) *QueryExecutor {
	return &QueryExecutor{pool: pool}
}

// ExecContext executes a query without using prepared statement cache
// Use this for dynamic queries that might change result types
func (qe *QueryExecutor) ExecContext(ctx context.Context, sql string, args ...interface{}) error {
	conn, err := qe.pool.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	// Use QueryExecModeExec to bypass prepared statement cache
	_, err = conn.Exec(ctx, sql, args...)
	return err
}

// QueryContext executes a query and returns rows without using prepared statement cache
func (qe *QueryExecutor) QueryContext(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	conn, err := qe.pool.Acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire connection: %w", err)
	}
	// Note: caller is responsible for calling rows.Close() which will release the connection

	// Use Query which bypasses prepared statement cache for dynamic queries
	return conn.Query(ctx, sql, args...)
}

// QueryRowContext executes a query that returns a single row without caching
func (qe *QueryExecutor) QueryRowContext(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	// QueryRow automatically handles connection acquisition and release
	return qe.pool.QueryRow(ctx, sql, args...)
}
