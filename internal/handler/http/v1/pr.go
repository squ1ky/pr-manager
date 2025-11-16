package v1

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/service"
)

type PullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type PullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"` // merged_at?
}

type PullRequestHandler struct {
	prSvc *service.PullRequestService
}

func NewPullRequestHandler(prSvc *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{prSvc: prSvc}
}

type createPRRequest struct {
	PullRequestID   string `json:"pull_request_id" binding:"required,notblank,max=255"`
	PullRequestName string `json:"pull_request_name" binding:"required,notblank,max=255"`
	AuthorID        string `json:"author_id" binding:"required,notblank,max=255"`
}

func (h *PullRequestHandler) Create(c *gin.Context) {
	var req createPRRequest
	if !bindAndValidate(c, &req) {
		return
	}

	pr, reviewerIDs, err := h.prSvc.Create(
		c.Request.Context(),
		req.PullRequestID,
		req.PullRequestName,
		req.AuthorID,
	)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyExists):
			respondError(c, http.StatusConflict, "PR_EXISTS", "PR id already exists")
		case errors.Is(err, repository.ErrNotFound):
			notFound(c, "resource not found")
		default:
			internalError(c)
		}
		return
	}

	resp := toPullRequestDTO(pr, reviewerIDs)

	c.JSON(http.StatusCreated, gin.H{
		"pr": resp,
	})
}

type mergePRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required,notblank,max=255"`
}

func (h *PullRequestHandler) Merge(c *gin.Context) {
	var req mergePRRequest
	if !bindAndValidate(c, &req) {
		return
	}

	pr, reviewerIDs, err := h.prSvc.Merge(c.Request.Context(), req.PullRequestID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			notFound(c, "resource not found")
		default:
			internalError(c)
		}
		return
	}

	resp := toPullRequestDTO(pr, reviewerIDs)

	c.JSON(http.StatusOK, gin.H{
		"pr": resp,
	})
}

type reassignPRRequest struct {
	PullRequestID string `json:"pull_request_id" binding:"required,notblank,max=255"`
	OldUserID     string `json:"old_user_id" binding:"required,notblank,max=255"`
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var req reassignPRRequest
	if !bindAndValidate(c, &req) {
		return
	}

	res, err := h.prSvc.Reassign(c.Request.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			notFound(c, "resource not found")
		case errors.Is(err, service.ErrPRMerged):
			respondError(c, http.StatusConflict, "PR_MERGED", "cannot reassign on merged PR")
		case errors.Is(err, service.ErrNotAssigned):
			respondError(c, http.StatusConflict, "NOT_ASSIGNED", "reviewer is not assigned to this PR")
		case errors.Is(err, service.ErrNoCandidate):
			respondError(c, http.StatusConflict, "NO_CANDIDATE", "no active replacement candidate in team")
		default:
			internalError(c)
		}
		return
	}

	resp := struct {
		PR         PullRequestDTO `json:"pr"`
		ReplacedBy string         `json:"replaced_by"`
	}{
		PR:         toPullRequestDTO(res.PR, res.ReviewerIDs),
		ReplacedBy: res.ReplacedByID,
	}

	c.JSON(http.StatusOK, resp)
}
