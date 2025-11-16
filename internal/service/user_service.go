package service

import (
	"context"

	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

// UserService provides operations for working with users.
type UserService struct {
	userRepo repository.UserRepository
}

// NewUserService constructs a new UserService with the given user repository.
func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{
		userRepo: users,
	}
}

// SetIsActive updates the isActive flag for the given user.
func (s *UserService) SetIsActive(ctx context.Context, userID string, active bool) (*entity.User, error) {
	return s.userRepo.SetUserActive(ctx, userID, active)
}

// GetByID returns a user by its identifier.
func (s *UserService) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}
