package repository

import (
	"context"
	"github.com/squ1ky/pr-manager/internal/entity"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, teamName string) error
	GetTeamByName(ctx context.Context, teamName string) (*entity.Team, []entity.User, error)
}

type UserRepository interface {
	UpsertUsers(ctx context.Context, users []entity.User) error
	GetUserByID(ctx context.Context, id string) (*entity.User, error)
	SetUserActive(ctx context.Context, userID string, active bool) (*entity.User, error)
	GetUsersByTeamName(ctx context.Context, teamName string) ([]entity.User, error)
}

type PullRequestRepository interface {
	CreatePullRequest(ctx context.Context, pr *entity.PullRequest) error
	GetPullRequestByID(ctx context.Context, id string) (*entity.PullRequest, error)
	GetPullRequestsByIDs(ctx context.Context, ids []string) ([]entity.PullRequest, error)
	MarkMerged(ctx context.Context, id string) (*entity.PullRequest, error)
}

type PullRequestReviewerRepository interface {
	SetForPullRequest(ctx context.Context, prID string, reviewerIDs []string) error
	GetByPullRequestID(ctx context.Context, prID string) ([]entity.PullRequestReviewer, error)
	GetPullRequestIDsByReviewer(ctx context.Context, reviewerID string) ([]string, error)
	AddReviewer(ctx context.Context, prID string, reviewerID string) error
	RemoveReviewer(ctx context.Context, prID string, reviewerID string) error
	IsReviewerAssigned(ctx context.Context, prID string, reviewerID string) (bool, error)
}
