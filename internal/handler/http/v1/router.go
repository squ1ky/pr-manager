package v1

import (
	"github.com/gin-gonic/gin"

	"github.com/squ1ky/pr-manager/internal/service"
)

type Handlers struct {
	Team        *TeamHandler
	User        *UserHandler
	PullRequest *PullRequestHandler
}

func NewRouter(
	r *gin.Engine,
	teamService *service.TeamService,
	userService *service.UserService,
	prService *service.PullRequestService,
) *Handlers {
	h := &Handlers{
		Team:        NewTeamHandler(teamService),
		User:        NewUserHandler(userService, prService),
		PullRequest: NewPullRequestHandler(prService),
	}

	v1 := r.Group("/api/v1")

	// Teams
	v1.POST("/team/add", h.Team.AddTeam)
	v1.GET("/team/get", h.Team.GetTeam)

	// Users
	v1.POST("/users/setIsActive", h.User.SetIsActive)
	v1.GET("/users/getReview", h.User.GetReview)
	v1.GET("/users/assignments/stat", h.User.GetUserAssignmentsStat)

	// PullRequests
	v1.POST("/pullRequest/create", h.PullRequest.Create)
	v1.POST("/pullRequest/merge", h.PullRequest.Merge)
	v1.POST("/pullRequest/reassign", h.PullRequest.Reassign)

	// Health
	v1.GET("/health", Health)

	return h
}
