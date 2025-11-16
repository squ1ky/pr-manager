package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"

	"github.com/squ1ky/pr-manager/config"
)

type Database struct {
	*sqlx.DB
	logger *slog.Logger
}

func NewPostgresConnection(cfg *config.DatabaseConfig, logger *slog.Logger) (*Database, error) {
	db, err := sqlx.Open("postgres", cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(cfg.MaxLifetime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("successfully connected to PostgreSQL")

	if err := RunMigrations(db.DB, logger); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &Database{
		DB:     db,
		logger: logger,
	}, nil
}

func (d *Database) Close() error {
	d.logger.Info("closing database connection")
	return d.DB.Close()
}

func (d *Database) Ping(ctx context.Context) error {
	return d.PingContext(ctx)
}
