package mocks

import (
	"context"

	"github.com/stretchr/testify/mock"

	"github.com/squ1ky/pr-manager/internal/entity"
)

// UserRepositoryMock mocks repository.UserRepository.
type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) UpsertUsers(ctx context.Context, users []entity.User) error {
	args := m.Called(ctx, users)
	return args.Error(0)
}

func (m *UserRepositoryMock) GetUserByID(ctx context.Context, id string) (*entity.User, error) {
	args := m.Called(ctx, id)
	u, _ := args.Get(0).(*entity.User)
	return u, args.Error(1)
}

func (m *UserRepositoryMock) SetUserActive(ctx context.Context, userID string, active bool) (*entity.User, error) {
	args := m.Called(ctx, userID, active)
	u, _ := args.Get(0).(*entity.User)
	return u, args.Error(1)
}

func (m *UserRepositoryMock) GetUsersByTeamName(ctx context.Context, teamName string) ([]entity.User, error) {
	args := m.Called(ctx, teamName)
	users, _ := args.Get(0).([]entity.User)
	return users, args.Error(1)
}

// PullRequestRepositoryMock mocks repository.PullRequestRepository.
type PullRequestRepositoryMock struct {
	mock.Mock
}

func (m *PullRequestRepositoryMock) CreatePullRequest(ctx context.Context, pr *entity.PullRequest) error {
	args := m.Called(ctx, pr)
	return args.Error(0)
}

func (m *PullRequestRepositoryMock) GetPullRequestByID(ctx context.Context, id string) (*entity.PullRequest, error) {
	args := m.Called(ctx, id)
	pr, _ := args.Get(0).(*entity.PullRequest)
	return pr, args.Error(1)
}

func (m *PullRequestRepositoryMock) GetPullRequestsByIDs(ctx context.Context, ids []string) ([]entity.PullRequest, error) {
	args := m.Called(ctx, ids)
	prs, _ := args.Get(0).([]entity.PullRequest)
	return prs, args.Error(1)
}

func (m *PullRequestRepositoryMock) MarkMerged(ctx context.Context, id string) (*entity.PullRequest, error) {
	args := m.Called(ctx, id)
	pr, _ := args.Get(0).(*entity.PullRequest)
	return pr, args.Error(1)
}

// PullRequestReviewerRepositoryMock mocks repository.PullRequestReviewerRepository.
type PullRequestReviewerRepositoryMock struct {
	mock.Mock
}

func (m *PullRequestReviewerRepositoryMock) SetForPullRequest(ctx context.Context, prID string, reviewerIDs []string) error {
	args := m.Called(ctx, prID, reviewerIDs)
	return args.Error(0)
}

func (m *PullRequestReviewerRepositoryMock) GetByPullRequestID(ctx context.Context, prID string) ([]entity.PullRequestReviewer, error) {
	args := m.Called(ctx, prID)
	reviewers, _ := args.Get(0).([]entity.PullRequestReviewer)
	return reviewers, args.Error(1)
}

func (m *PullRequestReviewerRepositoryMock) GetPullRequestIDsByReviewer(ctx context.Context, reviewerID string) ([]string, error) {
	args := m.Called(ctx, reviewerID)
	ids, _ := args.Get(0).([]string)
	return ids, args.Error(1)
}

func (m *PullRequestReviewerRepositoryMock) AddReviewer(ctx context.Context, prID string, reviewerID string) error {
	args := m.Called(ctx, prID, reviewerID)
	return args.Error(0)
}

func (m *PullRequestReviewerRepositoryMock) RemoveReviewer(ctx context.Context, prID string, reviewerID string) error {
	args := m.Called(ctx, prID, reviewerID)
	return args.Error(0)
}

func (m *PullRequestReviewerRepositoryMock) IsReviewerAssigned(ctx context.Context, prID string, reviewerID string) (bool, error) {
	args := m.Called(ctx, prID, reviewerID)
	return args.Bool(0), args.Error(1)
}

// TeamRepositoryMock mocks repository.TeamRepository.
type TeamRepositoryMock struct {
	mock.Mock
}

func (m *TeamRepositoryMock) CreateTeam(ctx context.Context, teamName string) error {
	args := m.Called(ctx, teamName)
	return args.Error(0)
}

func (m *TeamRepositoryMock) GetTeamByName(ctx context.Context, teamName string) (*entity.Team, []entity.User, error) {
	args := m.Called(ctx, teamName)
	team, _ := args.Get(0).(*entity.Team)
	members, _ := args.Get(1).([]entity.User)
	return team, members, args.Error(2)
}
