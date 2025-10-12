package database

import (
	"context"
	"fmt"
	"time"

	"github.com/binance-live/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Database wraps the PostgreSQL connection pool
type Database struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// New creates a new database connection pool
func New(cfg *config.DatabaseConfig, logger *zap.Logger) (*Database, error) {
	
	// Build connection string
	dsn := cfg.GetDSN()

	// Configure connection pool
	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {

		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	// Configure connection pool settings
	poolConfig.MaxConns = int32(cfg.MaxConnections)
	poolConfig.MinConns = int32(cfg.MaxIdleConnections)
	poolConfig.MaxConnLifetime = time.Duration(cfg.ConnectionMaxLifetime) * time.Second
	poolConfig.MaxConnIdleTime = 30 * time.Second
	poolConfig.HealthCheckPeriod = 30 * time.Second

	// Create connection pool
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {

		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {

		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("Database connection established",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Database),
	)

	return &Database{
		Pool:   pool,
		logger: logger,
	}, nil
}

// Close closes the database connection pool
func (db *Database) Close() {

	db.Pool.Close()
	db.logger.Info("Database connection closed")
}

// RunMigrations executes the database migrations
func (db *Database) RunMigrations(ctx context.Context, migrationSQL string) error {

	db.logger.Info("Running database migrations")

	_, err := db.Pool.Exec(ctx, migrationSQL)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	db.logger.Info("Database migrations completed successfully")
	return nil
}

// HealthCheck checks if the database is accessible
func (db *Database) HealthCheck(ctx context.Context) error {
	
	return db.Pool.Ping(ctx)
}
