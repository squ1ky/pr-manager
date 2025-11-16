package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

var (
	ErrPRMerged    = errors.New("pull request already merged")
	ErrNotAssigned = errors.New("reviewer is not assigned to PR")
	ErrNoCandidate = errors.New("no active replacement candidate")
)

// PullRequestService provides operations for managing pull requests and their reviewers.
type PullRequestService struct {
	prRepo       repository.PullRequestRepository
	reviewerRepo repository.PullRequestReviewerRepository
	userRepo     repository.UserRepository

	rnd *rand.Rand
}

// NewPullRequestService constructs a new PullRequestService with the given repositories.
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

// Create creates a new pull request and assigns up to two reviewers from the author's team.
func (s *PullRequestService) Create(
	ctx context.Context,
	prID string,
	name string,
	authorID string,
) (*entity.PullRequest, []string, error) {
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

	candidates := selectReviewersForCreate(teamMembers, authorID)
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

// Merge marks the pull request as merged and returns its current reviewers.
func (s *PullRequestService) Merge(ctx context.Context, prID string) (*entity.PullRequest, []string, error) {
	pr, err := s.prRepo.MarkMerged(ctx, prID)
	if err != nil {
		return nil, nil, err
	}

	reviewers, err := s.reviewerRepo.GetByPullRequestID(ctx, prID)
	if err != nil {
		return nil, nil, err
	}

	ids := reviewersToIDs(reviewers)

	return pr, ids, nil
}

// ReassignResult holds data produced by a successful reassignment.
type ReassignResult struct {
	PR           *entity.PullRequest
	ReviewerIDs  []string
	ReplacedByID string
}

// Reassign replaces one reviewer with another eligible teammate.
func (s *PullRequestService) Reassign(
	ctx context.Context,
	prID string,
	oldReviewerID string,
) (*ReassignResult, error) {
	pr, oldReviewer, err := s.validateReassignPreconditions(ctx, prID, oldReviewerID)
	if err != nil {
		return nil, err
	}

	teamMembers, err := s.userRepo.GetUsersByTeamName(ctx, *oldReviewer.TeamName)
	if err != nil {
		return nil, err
	}

	currentReviewers, err := s.reviewerRepo.GetByPullRequestID(ctx, prID)
	if err != nil {
		return nil, err
	}

	assignedSet := make(map[string]struct{}, len(currentReviewers))
	for _, r := range currentReviewers {
		assignedSet[r.ReviewerID] = struct{}{}
	}
	candidates := selectReviewersForReassign(teamMembers, pr.AuthorID, oldReviewerID, assignedSet)
	if len(candidates) == 0 {
		return nil, ErrNoCandidate
	}

	newReviewerID := s.pickRandomReviewers(candidates, 1)[0]

	if err := s.reviewerRepo.RemoveReviewer(ctx, prID, oldReviewerID); err != nil {
		return nil, err
	}
	if err := s.reviewerRepo.AddReviewer(ctx, prID, newReviewerID); err != nil {
		return nil, err
	}

	updatedReviewers, err := s.reviewerRepo.GetByPullRequestID(ctx, prID)
	if err != nil {
		return nil, err
	}

	return &ReassignResult{
		PR:           pr,
		ReviewerIDs:  reviewersToIDs(updatedReviewers),
		ReplacedByID: newReviewerID,
	}, nil
}

// validateReassignPreconditions loads PR and old reviewer and checks basic rules.
func (s *PullRequestService) validateReassignPreconditions(
	ctx context.Context,
	prID string,
	oldReviewerID string,
) (*entity.PullRequest, *entity.User, error) {
	pr, err := s.prRepo.GetPullRequestByID(ctx, prID)
	if err != nil {
		return nil, nil, err
	}
	if pr.Status == entity.PullRequestStatusMerged {
		return nil, nil, ErrPRMerged
	}

	assigned, err := s.reviewerRepo.IsReviewerAssigned(ctx, prID, oldReviewerID)
	if err != nil {
		return nil, nil, err
	}
	if !assigned {
		return nil, nil, ErrNotAssigned
	}

	oldReviewer, err := s.userRepo.GetUserByID(ctx, oldReviewerID)
	if err != nil {
		return nil, nil, err
	}
	if oldReviewer.TeamName == nil {
		return nil, nil, repository.ErrNotFound
	}

	return pr, oldReviewer, nil
}

// GetByReviewer returns all pull requests where the given user is a reviewer.
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

// GetUserAssignmentsStat returns user's assignment statistics
func (s *PullRequestService) GetUserAssignmentsStat(ctx context.Context, userID string) (repository.UserAssignmentsStat, error) {
	st, err := s.reviewerRepo.GetUserAssignmentsStat(ctx, userID)
	if err != nil {
		return repository.UserAssignmentsStat{}, err
	}

	return st, nil
}

// pickRandomReviewers picks up to n random reviewer IDs from the given users.
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

// selectReviewersForCreate returns active teammates except the author.
func selectReviewersForCreate(team []entity.User, authorID string) []entity.User {
	res := make([]entity.User, 0, len(team))
	for _, u := range team {
		if !u.IsActive || u.ID == authorID {
			continue
		}
		res = append(res, u)
	}
	return res
}

// selectReviewersForReassign returns active teammates except the author.
func selectReviewersForReassign(
	team []entity.User,
	authorID string,
	oldReviewerID string,
	assigned map[string]struct{},
) []entity.User {
	res := make([]entity.User, 0, len(team))
	for _, u := range team {
		if !u.IsActive || u.ID == authorID || u.ID == oldReviewerID {
			continue
		}
		if _, already := assigned[u.ID]; already {
			continue
		}
		res = append(res, u)
	}
	return res
}

// reviewersToIDs extracts reviewer IDs from the slice of PullRequestReviewer.
func reviewersToIDs(reviewers []entity.PullRequestReviewer) []string {
	ids := make([]string, 0, len(reviewers))
	for _, r := range reviewers {
		ids = append(ids, r.ReviewerID)
	}
	return ids
}
