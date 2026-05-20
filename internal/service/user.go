package service

import (
	"context"
	"fmt"
	"ssubench/internal/domain"
	"ssubench/internal/repository"
)

type UserService struct {
	repo *repository.UserRepository
}

func NewUserService(repo *repository.UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) Get(ctx context.Context, userID int) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID)

	if err != nil {
		return nil, fmt.Errorf("UserService.Get: %w", err)
	}

	return user, nil
}

func (s *UserService) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	users, err := s.repo.List(ctx, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("UserService.List: %w", err)
	}

	return users, nil
}

func (s *UserService) Block(ctx context.Context, userID int) (*domain.User, error) {
	user, err := s.repo.ChangeStatus(ctx, userID, domain.StatusBlocked)

	if err != nil {
		return nil, fmt.Errorf("UserService.Block: %w", err)
	}

	return user, nil
}

func (s *UserService) Unblock(ctx context.Context, userID int) (*domain.User, error) {
	user, err := s.repo.ChangeStatus(ctx, userID, domain.StatusActive)

	if err != nil {
		return nil, fmt.Errorf("UserService.Block: %w", err)
	}

	return user, nil
}

func (s *UserService) SetBalance(ctx context.Context, userID, amount int) (*domain.User, error) {
	user, err := s.repo.SetBalance(ctx, userID, amount)

	if err != nil {
		return nil, fmt.Errorf("UserService.Block: %w", err)
	}

	return user, nil
}
