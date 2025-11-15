package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/service"
	"net/http"
	"time"
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
	PullRequestID   string `json:"pull_request_id" binding:"required"`
	PullRequestName string `json:"pull_request_name" binding:"required"`
	AuthorID        string `json:"author_id" binding:"required"`
}

func (h *PullRequestHandler) Create(c *gin.Context) {
	var req createPRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "BAD_REQUEST", err.Error())
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
	PullRequestID string `json:"pull_request_id" binding:"required"`
}

func (h *PullRequestHandler) Merge(c *gin.Context) {
	var req mergePRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "BAD_REQUEST", err.Error())
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
	PullRequestID string `json:"pull_request_id" binding:"required"`
	OldUserID     string `json:"old_user_id" binding:"required"`
}

func (h *PullRequestHandler) Reassign(c *gin.Context) {
	var req reassignPRRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "BAD_REQUEST", err.Error())
		return
	}

	pr, reviewerIDs, replacedBy, err := h.prSvc.Reassign(c.Request.Context(), req.PullRequestID, req.OldUserID)
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
		PR:         toPullRequestDTO(pr, reviewerIDs),
		ReplacedBy: replacedBy,
	}

	c.JSON(http.StatusOK, resp)
}

// move to mapper

func toPullRequestDTO(pr *entity.PullRequest, reviewers []string) PullRequestDTO {
	dto := PullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: reviewers,
	}

	dto.MergedAt = pr.MergedAt

	return dto
}
