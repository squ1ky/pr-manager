package postgres

import (
	"context"
	"database/sql"
	"errors"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

type PullRequestRepository struct {
	db *sqlx.DB
}

func NewPullRequestRepository(db *sqlx.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) CreatePullRequest(ctx context.Context, pr *entity.PullRequest) error {
	const q = `
		INSERT INTO pull_requests (id, name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.ExecContext(ctx, q,
		pr.ID,
		pr.Name,
		pr.AuthorID,
		pr.Status,
	)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *PullRequestRepository) GetPullRequestByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	const q = `
		SELECT id, name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE id = $1
	`

	var pr entity.PullRequest
	if err := r.db.GetContext(ctx, &pr, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &pr, nil
}

func (r *PullRequestRepository) MarkMerged(ctx context.Context, id string) (*entity.PullRequest, error) {
	const q = `
		UPDATE pull_requests
		SET status = $2
			merged_at = COALESCE(merged_at, NOW())
		WHERE id = $1
		RETURNING id, name, author_id, status, created_at, merged_at
	`

	var pr entity.PullRequest
	if err := r.db.GetContext(ctx, &pr, q, id, entity.PullRequestStatusMerged); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &pr, nil
}
