package pgrepo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/testutils/testdb"
)

func TestTeamRepository_CreateTeam_Success(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewTeamRepository(db)

	teamName := "backend"

	err := repo.CreateTeam(ctx, teamName)
	require.NoError(t, err)

	team, members, err := repo.GetTeamByName(ctx, teamName)
	require.NoError(t, err)
	require.NotNil(t, team)
	require.Equal(t, teamName, team.Name)
	require.Len(t, members, 0)
}

func TestTeamRepository_CreateTeam_AlreadyExists(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewTeamRepository(db)

	teamName := "backend"

	err := repo.CreateTeam(ctx, teamName)
	require.NoError(t, err)

	err = repo.CreateTeam(ctx, teamName)
	require.Error(t, err)
	require.ErrorIs(t, err, repository.ErrAlreadyExists)
}

func TestTeamRepository_GetTeamByName_WithMembers(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	teamRepo := NewTeamRepository(db)
	userRepo := NewUserRepository(db)

	teamName := "backend"

	err := teamRepo.CreateTeam(ctx, teamName)
	require.NoError(t, err)

	users := []entity.User{
		{ID: "u1", Username: "alice", TeamName: &teamName, IsActive: true},
		{ID: "u2", Username: "bob", TeamName: &teamName, IsActive: false},
	}

	err = userRepo.UpsertUsers(ctx, users)
	require.NoError(t, err)

	team, members, err := teamRepo.GetTeamByName(ctx, teamName)
	require.NoError(t, err)
	require.NotNil(t, team)
	require.Equal(t, teamName, team.Name)

	require.Len(t, members, 2)
	require.ElementsMatch(t,
		[]string{"u1", "u2"},
		[]string{members[0].ID, members[1].ID},
	)
}

func TestTeamRepository_GetTeamByName_NotFound(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewTeamRepository(db)

	missingTeamName := "non-existent"

	team, members, err := repo.GetTeamByName(ctx, missingTeamName)
	require.Error(t, err)
	require.Nil(t, team)
	require.Nil(t, members)
	require.ErrorIs(t, err, repository.ErrNotFound)
}
