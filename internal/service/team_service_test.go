package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/mocks"
	"github.com/squ1ky/pr-manager/internal/repository"
)

type teamServiceTestDeps struct {
	ctx      context.Context
	teamRepo *mocks.TeamRepositoryMock
	userRepo *mocks.UserRepositoryMock
	svc      *TeamService
	tx       repository.TxManager
}

func setupTeamServiceTest(t *testing.T) teamServiceTestDeps {
	t.Helper()

	teamRepo := &mocks.TeamRepositoryMock{}
	userRepo := &mocks.UserRepositoryMock{}
	tx := noopTxManager{}

	svc := NewTeamService(teamRepo, userRepo, tx)

	return teamServiceTestDeps{
		ctx:      context.Background(),
		teamRepo: teamRepo,
		userRepo: userRepo,
		svc:      svc,
	}
}

// --- CreateTeamWithMembers ---

func TestTeamService_CreateTeamWithMembers_Success(t *testing.T) {
	deps := setupTeamServiceTest(t)

	teamName := "backend"

	members := []entity.User{
		{ID: "u1", Username: "alice"},
		{ID: "u2", Username: "bob"},
	}

	createdTeam := &entity.Team{
		Name: teamName,
	}
	actualMembers := []entity.User{
		{ID: "u1", Username: "alice", TeamName: &teamName},
		{ID: "u2", Username: "bob", TeamName: &teamName},
	}

	deps.teamRepo.
		On("CreateTeam", mock.Anything, teamName).
		Return(nil).
		Once()

	deps.userRepo.
		On("UpsertUsers",
			mock.Anything,
			mock.MatchedBy(func(users []entity.User) bool {
				if len(users) != len(members) {
					return false
				}
				for i, u := range users {
					if u.ID != members[i].ID || u.Username != members[i].Username {
						return false
					}
					if u.TeamName == nil || *u.TeamName != teamName {
						return false
					}
				}
				return true
			}),
		).
		Return(nil).
		Once()

	deps.teamRepo.
		On("GetTeamByName", mock.Anything, teamName).
		Return(createdTeam, actualMembers, nil).
		Once()

	team, gotMembers, err := deps.svc.CreateTeamWithMembers(deps.ctx, teamName, members)
	require.NoError(t, err)
	require.NotNil(t, team)

	assert.Equal(t, createdTeam, team)
	assert.Equal(t, actualMembers, gotMembers)

	deps.teamRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}

func TestTeamService_CreateTeamWithMembers_CreateTeamError(t *testing.T) {
	deps := setupTeamServiceTest(t)

	teamName := "backend"
	members := []entity.User{{ID: "u1"}}

	deps.teamRepo.
		On("CreateTeam", mock.Anything, teamName).
		Return(repository.ErrAlreadyExists).
		Once()

	team, gotMembers, err := deps.svc.CreateTeamWithMembers(deps.ctx, teamName, members)
	require.Error(t, err)
	assert.Nil(t, team)
	assert.Nil(t, gotMembers)
	assert.ErrorIs(t, err, repository.ErrAlreadyExists)

	deps.teamRepo.AssertExpectations(t)
	deps.userRepo.AssertNotCalled(t, "UpsertUsers", mock.Anything, mock.Anything)
}

func TestTeamService_CreateTeamWithMembers_UpsertError(t *testing.T) {
	deps := setupTeamServiceTest(t)

	teamName := "backend"
	members := []entity.User{{ID: "u1"}}

	deps.teamRepo.
		On("CreateTeam", mock.Anything, teamName).
		Return(nil).
		Once()

	deps.userRepo.
		On("UpsertUsers", mock.Anything, mock.AnythingOfType("[]entity.User")).
		Return(assert.AnError).
		Once()

	team, gotMembers, err := deps.svc.CreateTeamWithMembers(deps.ctx, teamName, members)
	require.Error(t, err)
	assert.Nil(t, team)
	assert.Nil(t, gotMembers)
	assert.ErrorIs(t, err, assert.AnError)

	deps.teamRepo.AssertExpectations(t)
	deps.userRepo.AssertExpectations(t)
}

// --- GetTeam ---

func TestTeamService_GetTeam_Success(t *testing.T) {
	deps := setupTeamServiceTest(t)

	teamName := "backend"

	team := &entity.Team{Name: teamName}
	members := []entity.User{
		{ID: "u1", Username: "alice", TeamName: &teamName},
	}

	deps.teamRepo.
		On("GetTeamByName", mock.Anything, teamName).
		Return(team, members, nil).
		Once()

	gotTeam, gotMembers, err := deps.svc.GetTeam(deps.ctx, teamName)
	require.NoError(t, err)
	assert.Equal(t, team, gotTeam)
	assert.Equal(t, members, gotMembers)

	deps.teamRepo.AssertExpectations(t)
}

func TestTeamService_GetTeam_NotFound(t *testing.T) {
	deps := setupTeamServiceTest(t)

	teamName := "backend"

	deps.teamRepo.
		On("GetTeamByName", mock.Anything, teamName).
		Return((*entity.Team)(nil), nil, repository.ErrNotFound).
		Once()

	gotTeam, gotMembers, err := deps.svc.GetTeam(deps.ctx, teamName)
	require.Error(t, err)
	assert.Nil(t, gotTeam)
	assert.Nil(t, gotMembers)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	deps.teamRepo.AssertExpectations(t)
}
