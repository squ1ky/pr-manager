package service

import (
	"context"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

// TeamService provides operations for managing teams and their members.
type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
}

// NewTeamService constructs a new TeamService with the given repositories.
func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
	}
}

// CreateTeamWithMembers creates a team and upserts its members into that team.
func (s *TeamService) CreateTeamWithMembers(ctx context.Context, teamName string, members []entity.User) (*entity.Team, []entity.User, error) {
	if err := s.teamRepo.CreateTeam(ctx, teamName); err != nil {
		return nil, nil, err
	}

	membersWithTeam := make([]entity.User, 0, len(members))
	for i := range members {
		u := members[i]
		u.TeamName = &teamName
		membersWithTeam = append(membersWithTeam, u)
	}

	if err := s.userRepo.UpsertUsers(ctx, membersWithTeam); err != nil {
		return nil, nil, err
	}

	team, actualMembers, err := s.teamRepo.GetTeamByName(ctx, teamName)
	if err != nil {
		return nil, nil, err
	}

	return team, actualMembers, nil
}

// GetTeam returns a team and its members by team name.
func (s *TeamService) GetTeam(ctx context.Context, teamName string) (*entity.Team, []entity.User, error) {
	return s.teamRepo.GetTeamByName(ctx, teamName)
}
