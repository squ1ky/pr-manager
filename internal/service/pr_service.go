package service

import (
	"context"
	"errors"
	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
	"math/rand"
	"time"
)

var (
	ErrPRMerged    = errors.New("pull request already merged")
	ErrNotAssigned = errors.New("reviewer is not assigned to PR")
	ErrNoCandidate = errors.New("no active replacement candidate")
)

type PullRequestService struct {
	prRepo       repository.PullRequestRepository
	reviewerRepo repository.PullRequestReviewerRepository
	userRepo     repository.UserRepository

	rnd *rand.Rand
}

func NewPullRequestService(
	prRepo repository.PullRequestRepository,
	reviewerRepo repository.PullRequestReviewerRepository,
	userRepo repository.UserRepository,
) *PullRequestService {
	return &PullRequestService{
		prRepo:       prRepo,
		reviewerRepo: reviewerRepo,
		userRepo:     userRepo,
		rnd:          rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (s *PullRequestService) Create(ctx context.Context, prID string, name string, authorID string) (*entity.PullRequest, []string, error) {
	author, err := s.userRepo.GetUserByID(ctx, authorID)
	if err != nil {
		return nil, nil, err
	}
	if author.TeamName == nil {
		return nil, nil, repository.ErrNotFound
	}

	teamMembers, err := s.userRepo.GetUsersByTeamName(ctx, *author.TeamName)
	if err != nil {
		return nil, nil, err
	}

	candidates := make([]entity.User, 0, len(teamMembers))
	for _, u := range teamMembers {
		if !u.IsActive {
			continue
		}
		if u.ID == author.ID {
			continue
		}
		candidates = append(candidates, u)
	}

	reviewerIDs := s.pickRandomReviewers(candidates, 2)
	pr := &entity.PullRequest{
		ID:       prID,
		Name:     name,
		AuthorID: authorID,
		Status:   entity.PullRequestStatusOpen,
	}

	if err := s.prRepo.CreatePullRequest(ctx, pr); err != nil {
		return nil, nil, err
	}

	if len(reviewerIDs) > 0 {
		if err := s.reviewerRepo.SetForPullRequest(ctx, pr.ID, reviewerIDs); err != nil {
			return nil, nil, err
		}
	}

	return pr, reviewerIDs, nil
}

func (s *PullRequestService) Merge(ctx context.Context, prID string) (*entity.PullRequest, []string, error) {
	pr, err := s.prRepo.MarkMerged(ctx, prID)
	if err != nil {
		return nil, nil, err
	}

	reviewers, err := s.reviewerRepo.GetByPullRequestID(ctx, prID)
	if err != nil {
		return nil, nil, err
	}

	ids := make([]string, 0, len(reviewers))
	for _, r := range reviewers {
		ids = append(ids, r.ReviewerID)
	}

	return pr, ids, nil
}

func (s *PullRequestService) Reassign(ctx context.Context, prID string, oldReviewerID string) (*entity.PullRequest, []string, string, error) {
	pr, err := s.prRepo.GetPullRequestByID(ctx, prID)
	if err != nil {
		return nil, nil, "", err
	}
	if pr.Status == entity.PullRequestStatusMerged {
		return nil, nil, "", ErrPRMerged
	}

	assigned, err := s.reviewerRepo.IsReviewerAssigned(ctx, prID, oldReviewerID)
	if err != nil {
		return nil, nil, "", err
	}
	if !assigned {
		return nil, nil, "", ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		return nil, nil, "", err
	}
	if oldReviewer.TeamName == nil {
		return nil, nil, "", repository.ErrNotFound
	}

	teamMembers, err := s.userRepo.GetUsersByTeamName(ctx, *oldReviewer.TeamName)
	if err != nil {
		return nil, nil, "", err
	}

	currentReviewers, err := s.reviewerRepo.GetByPullRequestID(ctx, prID)
	if err != nil {
		return nil, nil, "", err
	}
	assignedSet := make(map[string]struct{}, len(currentReviewers))
	for _, r := range currentReviewers {
		assignedSet[r.ReviewerID] = struct{}{}
	}

	candidates := make([]entity.User, 0, len(teamMembers))
	for _, u := range teamMembers {
		if !u.IsActive {
			continue
		}
		if u.ID == oldReviewerID {
			continue
		}
		if u.ID == pr.AuthorID {
			continue
		}
		if _, already := assignedSet[u.ID]; already {
			continue
		}
		candidates = append(candidates, u)
	}

	if len(candidates) == 0 {
		return nil, nil, "", ErrNoCandidate
	}

	newReviewer := s.pickRandomReviewers(candidates, 1)[0]

	if err := s.reviewerRepo.RemoveReviewer(ctx, prID, oldReviewerID); err != nil {
		return nil, nil, "", err
	}
	if err := s.reviewerRepo.AddReviewer(ctx, prID, newReviewer); err != nil {
		return nil, nil, "", err
	}

	updated, err := s.reviewerRepo.GetByPullRequestID(ctx, prID)
	if err != nil {
		return nil, nil, "", err
	}
	ids := make([]string, 0, len(updated))
	for _, r := range updated {
		ids = append(ids, r.ReviewerID)
	}

	return pr, ids, newReviewer, nil
}

func (s *PullRequestService) GetByReviewer(ctx context.Context, reviewerID string) ([]entity.PullRequest, error) {
	ids, err := s.reviewerRepo.GetPullRequestIDsByReviewer(ctx, reviewerID)
	if err != nil {
		return nil, err
	}
	if len(ids) == 0 {
		return []entity.PullRequest{}, nil
	}

	prs, err := s.prRepo.GetPullRequestsByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (s *PullRequestService) pickRandomReviewers(candidates []entity.User, n int) []string {
	if len(candidates) == 0 || n <= 0 {
		return nil
	}
	if len(candidates) <= n {
		ids := make([]string, 0, len(candidates))
		for _, c := range candidates {
			ids = append(ids, c.ID)
		}
		return ids
	}

	s.rnd.Shuffle(len(candidates), func(i, j int) {
		candidates[i], candidates[j] = candidates[j], candidates[i]
	})

	ids := make([]string, 0, n)
	for i := 0; i < n; i++ {
		ids = append(ids, candidates[i].ID)
	}
	return ids
}
