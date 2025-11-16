package pgrepo

import (
	"context"
	"testing"

	"github.com/jmoiron/sqlx"

	"github.com/stretchr/testify/require"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/testutils/testdb"
)

func createAuthorForTest(t *testing.T, db *sqlx.DB, id, username string) {
	t.Helper()

	ctx := context.Background()
	userRepo := NewUserRepository(db)

	user := entity.User{
		ID:       id,
		Username: username,
		TeamName: nil,
		IsActive: true,
	}

	err := userRepo.UpsertUsers(ctx, []entity.User{user})
	require.NoError(t, err)
}

func TestPullRequestRepository_CreatePullRequest_Success(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewPullRequestRepository(db)

	authorID := "author-1"
	authorName := "alice"
	createAuthorForTest(t, db, authorID, authorName)

	prID := "pr-1"
	prName := "Add feature X"

	pr := &entity.PullRequest{
		ID:       prID,
		Name:     prName,
		AuthorID: authorID,
		Status:   entity.PullRequestStatusOpen,
	}

	err := repo.CreatePullRequest(ctx, pr)
	require.NoError(t, err)

	got, err := repo.GetPullRequestByID(ctx, prID)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, prID, got.ID)
	require.Equal(t, prName, got.Name)
	require.Equal(t, authorID, got.AuthorID)
	require.Equal(t, entity.PullRequestStatusOpen, got.Status)
	require.False(t, got.CreatedAt.IsZero())
}

func TestPullRequestRepository_CreatePullRequest_AlreadyExists(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewPullRequestRepository(db)

	authorID := "author-2"
	authorName := "bob"
	createAuthorForTest(t, db, authorID, authorName)

	prID := "pr-dup"
	prName := "Duplicate PR"

	pr := &entity.PullRequest{
		ID:       prID,
		Name:     prName,
		AuthorID: authorID,
		Status:   entity.PullRequestStatusOpen,
	}

	err := repo.CreatePullRequest(ctx, pr)
	require.NoError(t, err)

	err = repo.CreatePullRequest(ctx, pr)
	require.Error(t, err)
	require.ErrorIs(t, err, repository.ErrAlreadyExists)
}

func TestPullRequestRepository_GetPullRequestByID_NotFound(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewPullRequestRepository(db)

	pr, err := repo.GetPullRequestByID(ctx, "non-existent-pr")
	require.Error(t, err)
	require.Nil(t, pr)
	require.ErrorIs(t, err, repository.ErrNotFound)
}

func TestPullRequestRepository_GetPullRequestsByIDs_EmptySlice(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewPullRequestRepository(db)

	prs, err := repo.GetPullRequestsByIDs(ctx, []string{})
	require.NoError(t, err)
	require.Len(t, prs, 0)
}

func TestPullRequestRepository_GetPullRequestsByIDs_Found(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewPullRequestRepository(db)

	authorID := "author-3"
	authorName := "carol"
	createAuthorForTest(t, db, authorID, authorName)

	prsToCreate := []*entity.PullRequest{
		{ID: "pr-10", Name: "PR 10", AuthorID: authorID, Status: entity.PullRequestStatusOpen},
		{ID: "pr-11", Name: "PR 11", AuthorID: authorID, Status: entity.PullRequestStatusOpen},
		{ID: "pr-12", Name: "PR 12", AuthorID: authorID, Status: entity.PullRequestStatusOpen},
	}

	for _, pr := range prsToCreate {
		err := repo.CreatePullRequest(ctx, pr)
		require.NoError(t, err)
	}

	ids := []string{"pr-10", "pr-12"}
	prs, err := repo.GetPullRequestsByIDs(ctx, ids)
	require.NoError(t, err)
	require.Len(t, prs, 2)

	gotIDs := []string{prs[0].ID, prs[1].ID}
	require.ElementsMatch(t, ids, gotIDs)
}

func TestPullRequestRepository_MarkMerged_Success(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewPullRequestRepository(db)

	authorID := "author-4"
	authorName := "dave"
	createAuthorForTest(t, db, authorID, authorName)

	prID := "pr-merge"
	pr := &entity.PullRequest{
		ID:       prID,
		Name:     "To be merged",
		AuthorID: authorID,
		Status:   entity.PullRequestStatusOpen,
	}

	err := repo.CreatePullRequest(ctx, pr)
	require.NoError(t, err)

	merged, err := repo.MarkMerged(ctx, prID)
	require.NoError(t, err)
	require.NotNil(t, merged)
	require.Equal(t, prID, merged.ID)
	require.Equal(t, entity.PullRequestStatusMerged, merged.Status)

	mergedAgain, err := repo.MarkMerged(ctx, prID)
	require.NoError(t, err)
	require.Equal(t, entity.PullRequestStatusMerged, mergedAgain.Status)
}

func TestPullRequestRepository_MarkMerged_NotFound(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	repo := NewPullRequestRepository(db)

	pr, err := repo.MarkMerged(ctx, "non-existent-pr")
	require.Error(t, err)
	require.Nil(t, pr)
	require.ErrorIs(t, err, repository.ErrNotFound)
}
