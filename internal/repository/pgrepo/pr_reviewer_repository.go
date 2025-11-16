package pgrepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

type PullRequestReviewerRepository struct {
	db *sqlx.DB
}

func NewPullRequestReviewerRepository(db *sqlx.DB) *PullRequestReviewerRepository {
	return &PullRequestReviewerRepository{db: db}
}

func (r *PullRequestReviewerRepository) SetForPullRequest(ctx context.Context, prID string, reviewerIDs []string) error {
	if tx, ok := txFromCtx(ctx); ok && tx != nil {
		return r.setForPullRequestWithExt(ctx, tx, prID, reviewerIDs)
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = r.setForPullRequestWithExt(ctx, tx, prID, reviewerIDs); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *PullRequestReviewerRepository) setForPullRequestWithExt(
	ctx context.Context,
	e extContext,
	prID string,
	reviewerIDs []string,
) (err error) {
	const deleteOld = `
       DELETE FROM pull_request_reviewers
       WHERE pull_request_id = $1
    `
	if _, err = e.ExecContext(ctx, deleteOld, prID); err != nil {
		return err
	}

	if len(reviewerIDs) > 0 {
		const insertNew = `
          INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
          VALUES ($1, $2)
        `
		for _, reviewerId := range reviewerIDs {
			if _, err = e.ExecContext(ctx, insertNew, prID, reviewerId); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *PullRequestReviewerRepository) GetByPullRequestID(ctx context.Context, prID string) ([]entity.PullRequestReviewer, error) {
	const q = `
		SELECT pull_request_id, reviewer_id
		FROM pull_request_reviewers
		WHERE pull_request_id = $1
	`
	e := ext(ctx, r.db)

	var reviewers []entity.PullRequestReviewer
	if err := sqlx.SelectContext(ctx, e, &reviewers, q, prID); err != nil {
		return nil, err
	}

	return reviewers, nil
}

func (r *PullRequestReviewerRepository) GetPullRequestIDsByReviewer(ctx context.Context, reviewerID string) ([]string, error) {
	const q = `
		SELECT pull_request_id
		FROM pull_request_reviewers
		WHERE reviewer_id = $1
	`
	e := ext(ctx, r.db)

	var ids []string
	if err := sqlx.SelectContext(ctx, e, &ids, q, reviewerID); err != nil {
		return nil, err
	}

	return ids, nil
}

func (r *PullRequestReviewerRepository) AddReviewer(ctx context.Context, prID string, reviewerID string) error {
	const q = `
		INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
		VALUES ($1, $2)
	`
	e := ext(ctx, r.db)

	if _, err := e.ExecContext(ctx, q, prID, reviewerID); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *PullRequestReviewerRepository) RemoveReviewer(ctx context.Context, prID string, reviewerID string) error {
	const q = `
		DELETE FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND reviewer_id = $2
	`
	e := ext(ctx, r.db)

	if _, err := e.ExecContext(ctx, q, prID, reviewerID); err != nil {
		return err
	}

	return nil
}

func (r *PullRequestReviewerRepository) IsReviewerAssigned(ctx context.Context, prID string, reviewerID string) (bool, error) {
	const q = `
		SELECT 1
		FROM pull_request_reviewers
		WHERE pull_request_id = $1 AND reviewer_id = $2
		LIMIT 1
	`
	e := ext(ctx, r.db)

	var dummy int
	if err := sqlx.GetContext(ctx, e, &dummy, q, prID, reviewerID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *PullRequestReviewerRepository) GetUserAssignmentsStat(ctx context.Context, userID string) (repository.UserAssignmentsStat, error) {
	const q = `
		SELECT COUNT(*) AS count
		FROM pull_request_reviewers
		WHERE reviewer_id = $1
	`
	e := ext(ctx, r.db)

	var count int
	if err := sqlx.GetContext(ctx, e, &count, q, userID); err != nil {
		return repository.UserAssignmentsStat{}, err
	}

	return repository.UserAssignmentsStat{
		UserID: userID,
		Count:  count,
	}, nil
}
