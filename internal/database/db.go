package database

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nghiavan0610/btaskee-quiz-service/config"
	"github.com/nghiavan0610/btaskee-quiz-service/internal/database/sqlc"
	"github.com/nghiavan0610/btaskee-quiz-service/pkg/logger"
)

type DatabaseConnection struct {
	Pool     *pgxpool.Pool
	Queries  *sqlc.Queries
	Executor *QueryExecutor
}

var (
	dbInstance *DatabaseConnection
	dbOnce     sync.Once
	dbError    error
)

func ProvideDatabasePool(db *DatabaseConnection) *pgxpool.Pool {
	return db.Pool
}

func ProvideDatabaseQueries(db *DatabaseConnection) *sqlc.Queries {
	return db.Queries
}

func ProvideDatabase(cfg *config.Config, log *logger.Logger) (*DatabaseConnection, error) {
	dbOnce.Do(func() {
		dbInstance, dbError = initializeDatabase(cfg, log)
	})
	return dbInstance, dbError
}

func initializeDatabase(cfg *config.Config, log *logger.Logger) (*DatabaseConnection, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?search_path=%s&sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
		cfg.Database.Schema,
		cfg.Database.SSLMode,
	)

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Error("Failed to parse database config", err)
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Pool settings for optimal performance
	poolConfig.MaxConns = int32(cfg.Database.MaxConnections)
	poolConfig.MinConns = int32(cfg.Database.MinConnections)
	poolConfig.MaxConnLifetime = time.Duration(cfg.Database.MaxConnLifetime) * time.Minute
	poolConfig.MaxConnIdleTime = time.Duration(cfg.Database.MaxConnIdleTime) * time.Minute

	// Disable prepared statement cache to avoid "cached plan must not change result type" errors
	// This is especially important for dynamic queries with conditional columns or complex CASE statements
	poolConfig.ConnConfig.DefaultQueryExecMode = pgx.QueryExecModeExec

	// Create connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Error("Failed to create database pool", err)
		return nil, fmt.Errorf("failed to create database pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		log.Error("Failed to ping database", err)
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	log.Info("[DATABASE] connection established",
		"host", cfg.Database.Host,
		"database", cfg.Database.DBName,
		"schema", cfg.Database.Schema,
		"max_conns", cfg.Database.MaxConnections,
	)

	// Create SQLC queries instance
	queries := sqlc.New(pool)

	// Create query executor for dynamic queries
	executor := NewQueryExecutor(pool)

	return &DatabaseConnection{
		Pool:     pool,
		Queries:  queries,
		Executor: executor,
	}, nil
}

func (db *DatabaseConnection) Close() {
	if db.Pool != nil {
		db.Pool.Close()
	}
}
