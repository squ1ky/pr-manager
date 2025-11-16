package pgrepo

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

type UserRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpsertUsers(ctx context.Context, users []entity.User) (err error) {
	if tx, ok := txFromCtx(ctx); ok && tx != nil {
		return r.upsertUsersWithExt(ctx, tx, users)
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

	if err = r.upsertUsersWithExt(ctx, tx, users); err != nil {
		return err
	}

	return tx.Commit()
}

func (r *UserRepository) upsertUsersWithExt(
	ctx context.Context,
	e extContext,
	users []entity.User,
) (err error) {
	const q = `
       INSERT INTO users (id, username, team_name, is_active)
       VALUES ($1, $2, $3, $4)
       ON CONFLICT (id) DO UPDATE
       SET username = EXCLUDED.username,
           team_name = EXCLUDED.team_name,
           is_active = EXCLUDED.is_active
    `
	for i := range users {
		u := &users[i]

		if _, err = e.ExecContext(ctx, q,
			u.ID,
			u.Username,
			u.TeamName,
			u.IsActive,
		); err != nil {
			return err
		}
	}

	return nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	const q = `
		SELECT id, username, team_name, is_active, created_at
    	FROM users
		WHERE id = $1
	`
	e := ext(ctx, r.db)

	var u entity.User
	if err := sqlx.GetContext(ctx, e, &u, q, id); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) SetUserActive(ctx context.Context, userID string, active bool) (*entity.User, error) {
	const q = `
		UPDATE users
		SET is_active = $2
		WHERE id = $1
		RETURNING id, username, team_name, is_active, created_at
	`
	e := ext(ctx, r.db)

	var u entity.User
	if err := sqlx.GetContext(ctx, e, &u, q, userID, active); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}

	return &u, nil
}

func (r *UserRepository) GetUsersByTeamName(ctx context.Context, teamName string) ([]entity.User, error) {
	const q = `
		SELECT id, username, team_name, is_active, created_at
		FROM users
		WHERE team_name = $1
	`
	e := ext(ctx, r.db)

	var users []entity.User
	if err := sqlx.SelectContext(ctx, e, &users, q, teamName); err != nil {
		return nil, err
	}

	return users, nil
}
