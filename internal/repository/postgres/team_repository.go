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

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, teamName string, members []entity.User) error {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	var teamID string
	const insertTeam = `
		INSERT INTO teams (name)
		VALUES ($1)
		RETURNING id
	`
	if err = tx.QueryRowContext(ctx, insertTeam, teamName).Scan(&teamID); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		return err
	}

	if len(members) > 0 {
		const updateUserTeam = `
			UPDATE users
			SET team_id = $2
			WHERE id = $1
		`
		for i := range members {
			m := members[i]
			if _, err = tx.ExecContext(ctx, updateUserTeam, m.ID, teamID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

func (r *TeamRepository) GetTeamByName(ctx context.Context, teamName string) (*entity.Team, []entity.User, error) {
	const selectTeam = `
		SELECT id, name, created_at
		FROM teams
		WHERE name = $1
	`

	var team entity.Team
	if err := r.db.GetContext(ctx, &team, selectTeam, teamName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, repository.ErrNotFound
		}
		return nil, nil, err
	}

	const selectMembers = `
		SELECT id, username, team_id, is_active, created_at
		FROM users
		WHERE team_id = $1
	`
	var members []entity.User
	if err := r.db.SelectContext(ctx, &members, selectMembers, team.ID); err != nil {
		return nil, nil, err
	}

	return &team, members, nil
}

func (r *TeamRepository) TeamExistsByName(ctx context.Context, teamName string) (bool, error) {
	const q = `
		SELECT 1
		FROM teams
		WHERE name = $1
		LIMIT 1
	`

	var dummy int
	if err := r.db.GetContext(ctx, &dummy, q, teamName); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
