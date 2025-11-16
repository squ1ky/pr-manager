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

type userServiceTestDeps struct {
	ctx      context.Context
	userRepo *mocks.UserRepositoryMock
	svc      *UserService
}

func setupUserServiceTest(t *testing.T) userServiceTestDeps {
	t.Helper()

	userRepo := &mocks.UserRepositoryMock{}
	svc := NewUserService(userRepo)

	return userServiceTestDeps{
		ctx:      context.Background(),
		userRepo: userRepo,
		svc:      svc,
	}
}

func TestUserService_GetByID_Success(t *testing.T) {
	deps := setupUserServiceTest(t)

	expected := &entity.User{
		ID:       "u1",
		Username: "test-user",
	}

	deps.userRepo.
		On("GetUserByID", mock.Anything, "u1").
		Return(expected, nil).
		Once()

	user, err := deps.svc.GetByID(deps.ctx, "u1")
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, expected, user)

	deps.userRepo.AssertExpectations(t)
}

func TestUserService_GetByID_NotFound(t *testing.T) {
	deps := setupUserServiceTest(t)

	deps.userRepo.
		On("GetUserByID", mock.Anything, "missing").
		Return((*entity.User)(nil), repository.ErrNotFound).
		Once()

	user, err := deps.svc.GetByID(deps.ctx, "missing")
	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	deps.userRepo.AssertExpectations(t)
}

func TestUserService_SetIsActive_Success(t *testing.T) {
	deps := setupUserServiceTest(t)

	updated := &entity.User{
		ID:       "u1",
		Username: "test-user",
		IsActive: true,
	}

	deps.userRepo.
		On("SetUserActive", mock.Anything, "u1", true).
		Return(updated, nil).
		Once()

	user, err := deps.svc.SetIsActive(deps.ctx, "u1", true)
	require.NoError(t, err)
	require.NotNil(t, user)
	assert.Equal(t, updated, user)

	deps.userRepo.AssertExpectations(t)
}

func TestUserService_SetIsActive_NotFound(t *testing.T) {
	deps := setupUserServiceTest(t)

	deps.userRepo.
		On("SetUserActive", mock.Anything, "missing", false).
		Return((*entity.User)(nil), repository.ErrNotFound).
		Once()

	user, err := deps.svc.SetIsActive(deps.ctx, "missing", false)
	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, repository.ErrNotFound)

	deps.userRepo.AssertExpectations(t)
}
