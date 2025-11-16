package service

import (
	"context"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/mocks"
	"github.com/squ1ky/pr-manager/internal/repository"
)

type prServiceTestDeps struct {
	ctx          context.Context
	prRepo       *mocks.PullRequestRepositoryMock
	reviewerRepo *mocks.PullRequestReviewerRepositoryMock
	userRepo     *mocks.UserRepositoryMock
	svc          *PullRequestService
}

func setupPRServiceTest(t *testing.T) prServiceTestDeps {
	t.Helper()

	prRepo := &mocks.PullRequestRepositoryMock{}
	reviewerRepo := &mocks.PullRequestReviewerRepositoryMock{}
	userRepo := &mocks.UserRepositoryMock{}

	svc := NewPullRequestService(prRepo, reviewerRepo, userRepo)
	svc.rnd = rand.New(rand.NewSource(1))

	return prServiceTestDeps{
		ctx:          context.Background(),
		prRepo:       prRepo,
		reviewerRepo: reviewerRepo,
		userRepo:     userRepo,
		svc:          svc,
	}
}

func assertPRMocks(t *testing.T, d prServiceTestDeps) {
	t.Helper()
	d.prRepo.AssertExpectations(t)
	d.reviewerRepo.AssertExpectations(t)
	d.userRepo.AssertExpectations(t)
}

// --- helpers ---

func TestSelectReviewersForCreate(t *testing.T) {
	team := []entity.User{
		{ID: "author", IsActive: true},
		{ID: "u1", IsActive: true},
		{ID: "u2", IsActive: false},
	}

	got := selectReviewersForCreate(team, "author")

	require.Len(t, got, 1)
	assert.Equal(t, "u1", got[0].ID)
}

func TestSelectReviewersForReassign(t *testing.T) {
	team := []entity.User{
		{ID: "author", IsActive: true},
		{ID: "old", IsActive: true},
		{ID: "u1", IsActive: true},
		{ID: "u2", IsActive: false},
	}

	assigned := map[string]struct{}{"u1": {}}

	got := selectReviewersForReassign(team, "author", "old", assigned)
	assert.Empty(t, got)

	delete(assigned, "u1")
	got = selectReviewersForReassign(team, "author", "old", assigned)

	require.Len(t, got, 1)
	assert.Equal(t, "u1", got[0].ID)
}

func TestReviewersToIDs(t *testing.T) {
	reviewers := []entity.PullRequestReviewer{
		{PullRequestID: "pr-1", ReviewerID: "u1"},
		{PullRequestID: "pr-1", ReviewerID: "u2"},
	}

	ids := reviewersToIDs(reviewers)

	assert.ElementsMatch(t, []string{"u1", "u2"}, ids)
}

// --- Create ---

func TestPullRequestService_Create_Success(t *testing.T) {
	deps := setupPRServiceTest(t)

	teamName := "team-a"

	author := &entity.User{
		ID:       "author",
		IsActive: true,
		TeamName: &teamName,
	}

	team := []entity.User{
		*author,
		{ID: "u1", IsActive: true, TeamName: &teamName},
		{ID: "u2", IsActive: false, TeamName: &teamName},
		{ID: "u3", IsActive: true, TeamName: &teamName},
	}

	deps.userRepo.
		On("GetUserByID", mock.Anything, "author").
		Return(author, nil).
		Once()

	deps.userRepo.
		On("GetUsersByTeamName", mock.Anything, teamName).
		Return(team, nil).
		Once()

	deps.prRepo.
		On("CreatePullRequest", mock.Anything, mock.AnythingOfType("*entity.PullRequest")).
		Return(nil).
		Once()

	deps.reviewerRepo.
		On("SetForPullRequest", mock.Anything, "pr-1", mock.AnythingOfType("[]string")).
		Return(nil).
		Once()

	pr, reviewers, err := deps.svc.Create(deps.ctx, "pr-1", "Test PR", "author")
	require.NoError(t, err)
	require.NotNil(t, pr)

	assert.Equal(t, "pr-1", pr.ID)
	assert.Equal(t, "Test PR", pr.Name)
	assert.Equal(t, "author", pr.AuthorID)
	assert.Equal(t, entity.PullRequestStatusOpen, pr.Status)

	require.LessOrEqual(t, len(reviewers), 2)
	for _, id := range reviewers {
		assert.NotEqual(t, "author", id)
	}

	assertPRMocks(t, deps)
}

func TestPullRequestService_Create_AuthorNotFound(t *testing.T) {
	deps := setupPRServiceTest(t)

	deps.userRepo.
		On("GetUserByID", mock.Anything, "author").
		Return((*entity.User)(nil), repository.ErrNotFound).
		Once()

	pr, reviewers, err := deps.svc.Create(deps.ctx, "pr-1", "Test PR", "author")
	require.Error(t, err)
	assert.Nil(t, pr)
	assert.Nil(t, reviewers)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	deps.userRepo.AssertExpectations(t)
}

func TestPullRequestService_Create_AuthorWithoutTeam(t *testing.T) {
	deps := setupPRServiceTest(t)

	author := &entity.User{
		ID:       "author",
		IsActive: true,
		TeamName: nil,
	}

	deps.userRepo.
		On("GetUserByID", mock.Anything, "author").
		Return(author, nil).
		Once()

	pr, reviewers, err := deps.svc.Create(deps.ctx, "pr-1", "Test PR", "author")
	require.Error(t, err)
	assert.Nil(t, pr)
	assert.Nil(t, reviewers)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	deps.userRepo.AssertExpectations(t)
}

// --- Merge ---

func TestPullRequestService_Merge_Success(t *testing.T) {
	deps := setupPRServiceTest(t)

	pr := &entity.PullRequest{
		ID:       "pr-1",
		Name:     "Test PR",
		AuthorID: "author",
		Status:   entity.PullRequestStatusOpen,
	}

	reviewers := []entity.PullRequestReviewer{
		{PullRequestID: pr.ID, ReviewerID: "u1"},
		{PullRequestID: pr.ID, ReviewerID: "u2"},
	}

	deps.prRepo.
		On("MarkMerged", mock.Anything, pr.ID).
		Return(pr, nil).
		Once()

	deps.reviewerRepo.
		On("GetByPullRequestID", mock.Anything, pr.ID).
		Return(reviewers, nil).
		Once()

	gotPR, ids, err := deps.svc.Merge(deps.ctx, pr.ID)
	require.NoError(t, err)
	assert.Equal(t, pr, gotPR)
	assert.ElementsMatch(t, []string{"u1", "u2"}, ids)

	deps.prRepo.AssertExpectations(t)
	deps.reviewerRepo.AssertExpectations(t)
}

// --- Reassign ---

func TestPullRequestService_Reassign_Success(t *testing.T) {
	deps := setupPRServiceTest(t)

	teamName := "team-a"

	pr := &entity.PullRequest{
		ID:       "pr-1",
		Name:     "Test PR",
		AuthorID: "author",
		Status:   entity.PullRequestStatusOpen,
	}

	author := &entity.User{ID: "author", IsActive: true, TeamName: &teamName}
	oldReviewer := &entity.User{ID: "old", IsActive: true, TeamName: &teamName}
	newReviewer := entity.User{ID: "new", IsActive: true, TeamName: &teamName}
	teamUsers := []entity.User{*author, *oldReviewer, newReviewer}

	currentReviewers := []entity.PullRequestReviewer{
		{PullRequestID: pr.ID, ReviewerID: "old"},
		{PullRequestID: pr.ID, ReviewerID: "other"},
	}
	updatedReviewers := []entity.PullRequestReviewer{
		{PullRequestID: pr.ID, ReviewerID: "new"},
		{PullRequestID: pr.ID, ReviewerID: "other"},
	}

	deps.prRepo.
		On("GetPullRequestByID", mock.Anything, pr.ID).
		Return(pr, nil).
		Once()

	deps.reviewerRepo.
		On("IsReviewerAssigned", mock.Anything, pr.ID, "old").
		Return(true, nil).
		Once()

	deps.userRepo.
		On("GetUserByID", mock.Anything, "old").
		Return(oldReviewer, nil).
		Once()

	deps.userRepo.
		On("GetUsersByTeamName", mock.Anything, teamName).
		Return(teamUsers, nil).
		Once()

	deps.reviewerRepo.
		On("GetByPullRequestID", mock.Anything, pr.ID).
		Return(currentReviewers, nil).
		Once()

	deps.reviewerRepo.
		On("RemoveReviewer", mock.Anything, pr.ID, "old").
		Return(nil).
		Once()

	deps.reviewerRepo.
		On("AddReviewer", mock.Anything, pr.ID, "new").
		Return(nil).
		Once()

	deps.reviewerRepo.
		On("GetByPullRequestID", mock.Anything, pr.ID).
		Return(updatedReviewers, nil).
		Once()

	res, err := deps.svc.Reassign(deps.ctx, pr.ID, "old")
	require.NoError(t, err)
	require.NotNil(t, res)

	assert.Equal(t, pr, res.PR)
	assert.Equal(t, "new", res.ReplacedByID)
	assert.ElementsMatch(t, []string{"new", "other"}, res.ReviewerIDs)

	assertPRMocks(t, deps)
}

func TestPullRequestService_Reassign_ErrPRMerged(t *testing.T) {
	deps := setupPRServiceTest(t)

	mergedPR := &entity.PullRequest{
		ID:       "pr-1",
		AuthorID: "author",
		Status:   entity.PullRequestStatusMerged,
	}

	deps.prRepo.
		On("GetPullRequestByID", mock.Anything, mergedPR.ID).
		Return(mergedPR, nil).
		Once()

	res, err := deps.svc.Reassign(deps.ctx, mergedPR.ID, "old")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrPRMerged)

	deps.prRepo.AssertExpectations(t)
}

func TestPullRequestService_Reassign_ErrNotAssigned(t *testing.T) {
	deps := setupPRServiceTest(t)

	pr := &entity.PullRequest{
		ID:       "pr-1",
		AuthorID: "author",
		Status:   entity.PullRequestStatusOpen,
	}

	deps.prRepo.
		On("GetPullRequestByID", mock.Anything, pr.ID).
		Return(pr, nil).
		Once()

	deps.reviewerRepo.
		On("IsReviewerAssigned", mock.Anything, pr.ID, "old").
		Return(false, nil).
		Once()

	res, err := deps.svc.Reassign(deps.ctx, pr.ID, "old")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrNotAssigned)

	deps.prRepo.AssertExpectations(t)
	deps.reviewerRepo.AssertExpectations(t)
}

func TestPullRequestService_Reassign_ErrNoCandidate(t *testing.T) {
	deps := setupPRServiceTest(t)

	teamName := "team-a"

	pr := &entity.PullRequest{
		ID:       "pr-1",
		AuthorID: "author",
		Status:   entity.PullRequestStatusOpen,
	}
	author := &entity.User{ID: "author", IsActive: true, TeamName: &teamName}
	oldReviewer := &entity.User{ID: "old", IsActive: true, TeamName: &teamName}

	deps.prRepo.
		On("GetPullRequestByID", mock.Anything, pr.ID).
		Return(pr, nil).
		Once()

	deps.reviewerRepo.
		On("IsReviewerAssigned", mock.Anything, pr.ID, "old").
		Return(true, nil).
		Once()

	deps.userRepo.
		On("GetUserByID", mock.Anything, "old").
		Return(oldReviewer, nil).
		Once()

	deps.userRepo.
		On("GetUsersByTeamName", mock.Anything, teamName).
		Return([]entity.User{*author, *oldReviewer}, nil).
		Once()

	deps.reviewerRepo.
		On("GetByPullRequestID", mock.Anything, pr.ID).
		Return([]entity.PullRequestReviewer{
			{PullRequestID: pr.ID, ReviewerID: "old"},
		}, nil).
		Once()

	res, err := deps.svc.Reassign(deps.ctx, pr.ID, "old")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.ErrorIs(t, err, ErrNoCandidate)

	deps.prRepo.AssertExpectations(t)
	deps.reviewerRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}

// --- GetByReviewer ---

func TestPullRequestService_GetByReviewer_Success(t *testing.T) {
	deps := setupPRServiceTest(t)

	ids := []string{"pr-1", "pr-2"}

	deps.reviewerRepo.
		On("GetPullRequestIDsByReviewer", mock.Anything, "u1").
		Return(ids, nil).
		Once()

	prs := []entity.PullRequest{
		{ID: "pr-1"},
		{ID: "pr-2"},
	}

	deps.prRepo.
		On("GetPullRequestsByIDs", mock.Anything, ids).
		Return(prs, nil).
		Once()

	res, err := deps.svc.GetByReviewer(deps.ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, prs, res)

	deps.prRepo.AssertExpectations(t)
	deps.reviewerRepo.AssertExpectations(t)
}

func TestPullRequestService_GetByReviewer_Empty(t *testing.T) {
	deps := setupPRServiceTest(t)

	deps.reviewerRepo.
		On("GetPullRequestIDsByReviewer", mock.Anything, "u1").
		Return([]string{}, nil).
		Once()

	res, err := deps.svc.GetByReviewer(deps.ctx, "u1")
	require.NoError(t, err)
	assert.Empty(t, res)

	deps.reviewerRepo.AssertExpectations(t)
}
