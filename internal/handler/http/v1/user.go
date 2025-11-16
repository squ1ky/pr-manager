package v1

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/service"
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
	UserID   string `json:"user_id" binding:"required,notblank,max=255"`
	IsActive *bool  `json:"is_active" binding:"required"`
}

func (h *UserHandler) SetIsActive(c *gin.Context) {
	var req setIsActiveRequest
	if !bindAndValidate(c, &req) {
		return
	}

	user, err := h.userSvc.SetIsActive(c.Request.Context(), req.UserID, *req.IsActive)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			notFound(c, "resource not found")
		default:
			internalError(c)
		}
		return
	}

	resp := toUserDTO(user)

	c.JSON(http.StatusOK, gin.H{
		"user": resp,
	})
}

type getReviewRequest struct {
	UserID string `form:"user_id" binding:"required,notblank,max=255"`
}

type getReviewResponse struct {
	UserID       string                `json:"user_id"`
	PullRequests []PullRequestShortDTO `json:"pull_requests"`
}

func (h *UserHandler) GetReview(c *gin.Context) {
	var req getReviewRequest
	if !bindAndValidate(c, &req) {
		return
	}

	if _, err := h.userSvc.GetByID(c.Request.Context(), req.UserID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			notFound(c, "resource not found")
		} else {
			internalError(c)
		}
		return
	}

	prs, err := h.prSvc.GetByReviewer(c.Request.Context(), req.UserID)
	if err != nil {
		internalError(c)
		return
	}

	resp := getReviewResponse{
		UserID:       req.UserID,
		PullRequests: toPullRequestShortList(prs),
	}

	c.JSON(http.StatusOK, resp)
}
