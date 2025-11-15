package service

import (
	"context"
	"github.com/squ1ky/pr-manager/internal/entity"
	"github.com/squ1ky/pr-manager/internal/repository"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{
		userRepo: users,
	}
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, active bool) (*entity.User, error) {
	return s.userRepo.SetUserActive(ctx, userID, active)
}

func (s *UserService) GetByID(ctx context.Context, userID string) (*entity.User, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}
