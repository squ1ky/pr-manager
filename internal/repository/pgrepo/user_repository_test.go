package pgrepo

import (
	"context"
	"testing"

	"github.com/squ1ky/pr-manager/internal/testutils/testdb"

	"github.com/stretchr/testify/require"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

func createTeamForTest(t *testing.T, name string) {
	t.Helper()

	db := testdb.GetDB(t)
	ctx := context.Background()

	_, err := db.ExecContext(ctx, `INSERT INTO teams (name) VALUES ($1)`, name)
	require.NoError(t, err)
}

func TestUserRepository_UpsertAndGetByID(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewUserRepository(db)

	backendTeam := "backend"

	createTeamForTest(t, backendTeam)

	user := entity.User{
		ID:       "user-1",
		Username: "alice",
		TeamName: &backendTeam,
		IsActive: true,
	}

	err := repo.UpsertUsers(ctx, []entity.User{user})
	require.NoError(t, err)

	got, err := repo.GetUserByID(ctx, user.ID)
	require.NoError(t, err)

	require.Equal(t, user.ID, got.ID)
	require.Equal(t, user.Username, got.Username)
	require.Equal(t, user.TeamName, got.TeamName)
	require.Equal(t, user.IsActive, got.IsActive)
	require.False(t, got.CreatedAt.IsZero())
}

func TestUserRepository_Upsert_UpdatesExisting(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewUserRepository(db)

	backendTeam := "backend"

	createTeamForTest(t, backendTeam)

	u := entity.User{
		ID:       "user-2",
		Username: "bob",
		TeamName: &backendTeam,
		IsActive: true,
	}

	require.NoError(t, repo.UpsertUsers(ctx, []entity.User{u}))

	u.Username = "bob-updated"
	u.IsActive = false

	require.NoError(t, repo.UpsertUsers(ctx, []entity.User{u}))

	got, err := repo.GetUserByID(ctx, u.ID)
	require.NoError(t, err)
	require.Equal(t, "bob-updated", got.Username)
	require.False(t, got.IsActive)
}

func TestUserRepository_SetUserActive(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewUserRepository(db)

	backendTeam := "backend"

	createTeamForTest(t, backendTeam)

	u := entity.User{
		ID:       "user-3",
		Username: "carol",
		TeamName: &backendTeam,
		IsActive: false,
	}

	require.NoError(t, repo.UpsertUsers(ctx, []entity.User{u}))

	updated, err := repo.SetUserActive(ctx, u.ID, true)
	require.NoError(t, err)
	require.True(t, updated.IsActive)

	got, err := repo.GetUserByID(ctx, u.ID)
	require.NoError(t, err)
	require.True(t, got.IsActive)
}

func TestUserRepository_GetUsersByTeamName(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewUserRepository(db)

	frontendTeam := "frontend"
	backendTeam := "backend"

	createTeamForTest(t, backendTeam)
	createTeamForTest(t, frontendTeam)

	users := []entity.User{
		{ID: "u1", Username: "alice", TeamName: &backendTeam, IsActive: true},
		{ID: "u2", Username: "bob", TeamName: &backendTeam, IsActive: true},
		{ID: "u3", Username: "dave", TeamName: &frontendTeam, IsActive: false},
	}

	require.NoError(t, repo.UpsertUsers(ctx, users))

	backendUsers, err := repo.GetUsersByTeamName(ctx, backendTeam)
	require.NoError(t, err)
	require.Len(t, backendUsers, 2)

	frontendUsers, err := repo.GetUsersByTeamName(ctx, frontendTeam)
	require.NoError(t, err)
	require.Len(t, frontendUsers, 1)
	require.Equal(t, "dave", frontendUsers[0].Username)
}

func TestUserRepository_GetUserByID_NotFound(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewUserRepository(db)

	u, err := repo.GetUserByID(ctx, "non-existent")
	require.Error(t, err)
	require.Nil(t, u)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestUserRepository_SetUserActive_NotFound(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewUserRepository(db)

	u, err := repo.SetUserActive(ctx, "non-existent", true)
	require.Error(t, err)
	require.Nil(t, u)
	require.ErrorIs(t, err, repository.ErrNotFound)
}
