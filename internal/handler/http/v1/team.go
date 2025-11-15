package v1

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
	"github.com/squ1ky/pr-manager/internal/service"
	"log/slog"
	"net/http"
)

type TeamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []TeamMemberDTO `json:"members"`
}

type TeamHandler struct {
	teamSvc *service.TeamService
}

func NewTeamHandler(teamSvc *service.TeamService) *TeamHandler {
	return &TeamHandler{teamSvc: teamSvc}
}

func (h *TeamHandler) AddTeam(c *gin.Context) {
	var req TeamDTO

	if err := c.ShouldBindJSON(&req); err != nil {
		badRequest(c, "BAD_REQUEST", err.Error())
		return
	}

	members := make([]entity.User, 0, len(req.Members))
	for _, m := range req.Members {
		members = append(members, entity.User{
			ID:       m.UserID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}

	team, actualMembers, err := h.teamSvc.CreateTeamWithMembers(c.Request.Context(), req.TeamName, members)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrAlreadyExists):
			badRequest(c, "TEAM_EXISTS", "team_name already exists")
		default:
			slog.Error("AddTeam failed", slog.String("error", err.Error()))
			internalError(c)
		}
		return
	}

	resp := TeamDTO{
		TeamName: team.Name,
		Members:  make([]TeamMemberDTO, 0, len(actualMembers)),
	}
	for _, u := range actualMembers {
		resp.Members = append(resp.Members, TeamMemberDTO{
			UserID:   u.ID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	c.JSON(http.StatusCreated, gin.H{
		"team": resp,
	})
}

func (h *TeamHandler) GetTeam(c *gin.Context) {
	teamName := c.Query("team_name")
	if teamName == "" {
		badRequest(c, "BAD_REQUEST", "team_name is required")
		return
	}

	team, members, err := h.teamSvc.GetTeam(c.Request.Context(), teamName)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFound):
			notFound(c, "resource not found")
		default:
			internalError(c)
		}
		return
	}

	resp := TeamDTO{
		TeamName: team.Name,
		Members:  make([]TeamMemberDTO, 0, len(members)),
	}
	for _, u := range members {
		resp.Members = append(resp.Members, TeamMemberDTO{
			UserID:   u.ID,
			Username: u.Username,
			IsActive: u.IsActive,
		})
	}

	c.JSON(http.StatusOK, resp)
}
