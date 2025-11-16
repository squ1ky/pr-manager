package testdb

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	once      sync.Once
	db        *sqlx.DB
	container testcontainers.Container
	initErr   error
)

func GetDB(t *testing.T) *sqlx.DB {
	t.Helper()

	once.Do(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()

		req := testcontainers.ContainerRequest{
			Image:        "postgres:16-alpine",
			ExposedPorts: []string{"5432/tcp"},
			Env: map[string]string{
				"POSTGRES_DB":       "prmanager_test",
				"POSTGRES_USER":     "test",
				"POSTGRES_PASSWORD": "test",
			},
			WaitingFor: wait.ForListeningPort("5432/tcp").
				WithStartupTimeout(60 * time.Second),
		}

		c, err := testcontainers.GenericContainer(
			ctx,
			testcontainers.GenericContainerRequest{
				ContainerRequest: req,
				Started:          true,
			},
		)
		if err != nil {
			initErr = fmt.Errorf("cannot start postgres container: %w", err)
			return
		}
		container = c

		host, err := c.Host(ctx)
		if err != nil {
			initErr = fmt.Errorf("cannot get container host: %w", err)
			return
		}

		mappedPort, err := c.MappedPort(ctx, "5432/tcp")
		if err != nil {
			initErr = fmt.Errorf("cannot get mapped port: %w", err)
			return
		}

		dsn := fmt.Sprintf(
			"postgres://%s:%s@%s:%s/%s?sslmode=disable",
			"test",
			"test",
			host,
			mappedPort.Port(),
			"prmanager_test",
		)

		db, err = sqlx.Open("postgres", dsn)
		if err != nil {
			initErr = fmt.Errorf("cannot open db: %w", err)
			return
		}

		if err := db.PingContext(ctx); err != nil {
			initErr = fmt.Errorf("cannot ping db: %w", err)
			return
		}

		if err := applySchema(ctx, db.DB); err != nil {
			initErr = fmt.Errorf("cannot apply schema: %w", err)
			return
		}
	})

	if initErr != nil {
		require.FailNow(t, "failed to init test db", "%v", initErr)
	}

	return db
}

func TruncateAll(t *testing.T) {
	t.Helper()

	require.NotNil(t, db, "db not initialized, call GetDB first")

	const q = `
		TRUNCATE TABLE
			pull_request_reviewers,
			pull_requests,
			users,
			teams
		RESTART IDENTITY CASCADE;
`

	ctx := context.Background()
	_, err := db.ExecContext(ctx, q)
	require.NoError(t, err)
}

func Teardown() {
	if container == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := container.Terminate(ctx); err != nil {
		log.Printf("failed to terminate postgres container: %v", err)
	}
}

func applySchema(ctx context.Context, db *sql.DB) error {
	const schema = `
CREATE TABLE IF NOT EXISTS teams
(
    name       VARCHAR(255) PRIMARY KEY,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS users
(
    id         VARCHAR(255) PRIMARY KEY,
    username   VARCHAR(255) NOT NULL,
    team_name  VARCHAR(255) REFERENCES teams (name) ON DELETE SET NULL,
    is_active  BOOLEAN   DEFAULT true,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS pull_requests
(
    id         VARCHAR(255) PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    author_id  VARCHAR(255) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    status     VARCHAR(20)  NOT NULL CHECK (status in ('OPEN', 'MERGED')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    merged_at  TIMESTAMP    NULL
);

CREATE TABLE IF NOT EXISTS pull_request_reviewers
(
    pull_request_id VARCHAR(255) NOT NULL REFERENCES pull_requests (id) ON DELETE CASCADE,
    reviewer_id     VARCHAR(255) NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    PRIMARY KEY (pull_request_id, reviewer_id)
);
`
	_, err := db.ExecContext(ctx, schema)
	return err
}
