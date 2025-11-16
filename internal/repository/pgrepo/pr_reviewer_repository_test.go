package pgrepo

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/testutils/testdb"
)

func createUserForTest(t *testing.T, ctx context.Context, repo *UserRepository, id, username string) {
	t.Helper()

	u := entity.User{
		ID:       id,
		Username: username,
		TeamName: nil,
		IsActive: true,
	}

	err := repo.UpsertUsers(ctx, []entity.User{u})
	require.NoError(t, err)
}

func createPullRequestForTest(t *testing.T, ctx context.Context, repo *PullRequestRepository, id, name, authorID string) {
	t.Helper()

	pr := &entity.PullRequest{
		ID:       id,
		Name:     name,
		AuthorID: authorID,
		Status:   entity.PullRequestStatusOpen,
	}

	err := repo.CreatePullRequest(ctx, pr)
	require.NoError(t, err)
}

func TestPullRequestReviewerRepository_SetForPullRequest_AddAndReplace(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	prRepo := NewPullRequestRepository(db)
	revRepo := NewPullRequestReviewerRepository(db)

	authorID := "author-1"
	authorName := "alice"
	createUserForTest(t, ctx, userRepo, authorID, authorName)

	prID := "pr-1"
	prName := "Feature A"
	createPullRequestForTest(t, ctx, prRepo, prID, prName, authorID)

	reviewer1ID := "rev-1"
	reviewer2ID := "rev-2"
	reviewer3ID := "rev-3"

	createUserForTest(t, ctx, userRepo, reviewer1ID, "bob")
	createUserForTest(t, ctx, userRepo, reviewer2ID, "carol")
	createUserForTest(t, ctx, userRepo, reviewer3ID, "dave")

	initialReviewers := []string{reviewer1ID, reviewer2ID}
	err := revRepo.SetForPullRequest(ctx, prID, initialReviewers)
	require.NoError(t, err)

	got, err := revRepo.GetByPullRequestID(ctx, prID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	gotIDs := []string{got[0].ReviewerID, got[1].ReviewerID}
	require.ElementsMatch(t, initialReviewers, gotIDs)

	updatedReviewers := []string{reviewer2ID, reviewer3ID}
	err = revRepo.SetForPullRequest(ctx, prID, updatedReviewers)
	require.NoError(t, err)

	got, err = revRepo.GetByPullRequestID(ctx, prID)
	require.NoError(t, err)
	require.Len(t, got, 2)
	gotIDs = []string{got[0].ReviewerID, got[1].ReviewerID}
	require.ElementsMatch(t, updatedReviewers, gotIDs)
}

func TestPullRequestReviewerRepository_SetForPullRequest_ClearAll(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	prRepo := NewPullRequestRepository(db)
	revRepo := NewPullRequestReviewerRepository(db)

	authorID := "author-2"
	authorName := "eve"
	createUserForTest(t, ctx, userRepo, authorID, authorName)

	prID := "pr-2"
	prName := "Feature B"
	createPullRequestForTest(t, ctx, prRepo, prID, prName, authorID)

	reviewerID := "rev-10"
	createUserForTest(t, ctx, userRepo, reviewerID, "frank")

	err := revRepo.SetForPullRequest(ctx, prID, []string{reviewerID})
	require.NoError(t, err)

	err = revRepo.SetForPullRequest(ctx, prID, []string{})
	require.NoError(t, err)

	got, err := revRepo.GetByPullRequestID(ctx, prID)
	require.NoError(t, err)
	require.Len(t, got, 0)
}

func TestPullRequestReviewerRepository_GetPullRequestIDsByReviewer(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	prRepo := NewPullRequestRepository(db)
	revRepo := NewPullRequestReviewerRepository(db)

	authorID := "author-3"
	authorName := "grace"
	createUserForTest(t, ctx, userRepo, authorID, authorName)

	reviewerID := "rev-20"
	createUserForTest(t, ctx, userRepo, reviewerID, "henry")

	pr1ID := "pr-3"
	pr2ID := "pr-4"
	createPullRequestForTest(t, ctx, prRepo, pr1ID, "PR 3", authorID)
	createPullRequestForTest(t, ctx, prRepo, pr2ID, "PR 4", authorID)

	require.NoError(t, revRepo.AddReviewer(ctx, pr1ID, reviewerID))
	require.NoError(t, revRepo.AddReviewer(ctx, pr2ID, reviewerID))

	ids, err := revRepo.GetPullRequestIDsByReviewer(ctx, reviewerID)
	require.NoError(t, err)
	require.Len(t, ids, 2)
	require.ElementsMatch(t, []string{pr1ID, pr2ID}, ids)

	anotherReviewerID := "rev-21"
	createUserForTest(t, ctx, userRepo, anotherReviewerID, "ivan")

	ids, err = revRepo.GetPullRequestIDsByReviewer(ctx, anotherReviewerID)
	require.NoError(t, err)
	require.Len(t, ids, 0)
}

func TestPullRequestReviewerRepository_AddReviewer_SuccessAndAlreadyExists(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	prRepo := NewPullRequestRepository(db)
	revRepo := NewPullRequestReviewerRepository(db)

	authorID := "author-4"
	authorName := "jack"
	createUserForTest(t, ctx, userRepo, authorID, authorName)

	reviewerID := "rev-30"
	createUserForTest(t, ctx, userRepo, reviewerID, "kate")

	prID := "pr-5"
	createPullRequestForTest(t, ctx, prRepo, prID, "PR 5", authorID)

	err := revRepo.AddReviewer(ctx, prID, reviewerID)
	require.NoError(t, err)

	assigned, err := revRepo.IsReviewerAssigned(ctx, prID, reviewerID)
	require.NoError(t, err)
	require.True(t, assigned)

	err = revRepo.AddReviewer(ctx, prID, reviewerID)
	require.Error(t, err)
	require.ErrorIs(t, err, repository.ErrAlreadyExists)
}

func TestPullRequestReviewerRepository_RemoveReviewer(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	userRepo := NewUserRepository(db)
	prRepo := NewPullRequestRepository(db)
	revRepo := NewPullRequestReviewerRepository(db)

	authorID := "author-5"
	authorName := "lisa"
	createUserForTest(t, ctx, userRepo, authorID, authorName)

	reviewerID := "rev-40"
	createUserForTest(t, ctx, userRepo, reviewerID, "mike")

	prID := "pr-6"
	createPullRequestForTest(t, ctx, prRepo, prID, "PR 6", authorID)

	require.NoError(t, revRepo.AddReviewer(ctx, prID, reviewerID))

	err := revRepo.RemoveReviewer(ctx, prID, reviewerID)
	require.NoError(t, err)

	assigned, err := revRepo.IsReviewerAssigned(ctx, prID, reviewerID)
	require.NoError(t, err)
	require.False(t, assigned)

	ids, err := revRepo.GetPullRequestIDsByReviewer(ctx, reviewerID)
	require.NoError(t, err)
	require.Len(t, ids, 0)
}

func TestPullRequestReviewerRepository_IsReviewerAssigned_NoRow(t *testing.T) {
	db := testdb.GetDB(t)
	testdb.TruncateAll(t)

	ctx := context.Background()
	revRepo := NewPullRequestReviewerRepository(db)

	prID := "pr-unknown"
	reviewerID := "rev-unknown"

	assigned, err := revRepo.IsReviewerAssigned(ctx, prID, reviewerID)
	require.NoError(t, err)
	require.False(t, assigned)
}
