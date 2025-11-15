package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/service"
	"net/http"
)

type UserDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type UserHandler struct {
	userSvc *service.UserService
	prSvc   *service.PullRequestService
}

func NewUserHandler(userSvc *service.UserService, prSvc *service.PullRequestService) *UserHandler {
	return &UserHandler{
		userSvc: userSvc,
		prSvc:   prSvc,
	}
}

type setIsActiveRequest struct {
	UserID   string `json:"user_id" binding:"required"`
	IsActive *bool  `json:"is_active" binding:"required"`
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req setIsActiveRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "BAD_REQUEST", err.Error())
		return
	}

	user, err := h.userSvc.SetIsActive(c.Request.Context(), req.UserID, *req.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			notFound(c, "user_not_found")
		default:
			internalError(c)
		}
		return
	}

	resp := UserDTO{
		UserID:   user.ID,
		Username: user.Username,
		TeamName: *user.TeamName,
		IsActive: user.IsActive,
	}

	c.JSON(http.StatusOK, gin.H{
		"user": resp,
	})
}

type getReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

func (h *UserHandler) GetReview(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		badRequest(c, "BAD_REQUEST", "user_id is required")
		return
	}

	if _, err := h.userSvc.GetByID(c.Request.Context(), userID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			notFound(c, "user_not_found")
		} else {
			internalError(c)
		}
		return
	}

	prs, err := h.prSvc.GetByReviewer(c.Request.Context(), userID)
	if err != nil {
		internalError(c)
		return
	}

	resp := getReviewResponse{
		UserID:       userID,
		PullRequests: make([]PullRequestShortDTO, 0, len(prs)),
	}

	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, PullRequestShortDTO{
			PullRequestID:   pr.ID,
			PullRequestName: pr.Name,
			AuthorID:        pr.AuthorID,
			Status:          string(pr.Status),
		})
	}

	c.JSON(http.StatusOK, resp)
}
