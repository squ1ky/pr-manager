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

type TeamRepository struct {
	db *sqlx.DB
}

func NewTeamRepository(db *sqlx.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, teamName string) error {
	const insertTeam = `
		INSERT INTO teams (name)
		VALUES ($1)
	`

	if _, err := r.db.ExecContext(ctx, insertTeam, teamName); err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		return err
	}

	return nil
}

func (r *TeamRepository) GetTeamByName(ctx context.Context, teamName string) (*entity.Team, []entity.User, error) {
	const selectTeam = `
		SELECT name, created_at
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
		SELECT id, username, team_name, is_active, created_at
		FROM users
		WHERE team_name = $1
	`
	var members []entity.User
	if err := r.db.SelectContext(ctx, &members, selectMembers, team.Name); err != nil {
		return nil, nil, err
	}

	return &team, members, nil
}
