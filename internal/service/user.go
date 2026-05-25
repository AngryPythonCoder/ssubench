package service

import (
	"context"
	"fmt"
	"ssubench/internal/domain"
)

type UserService struct {
	txManager domain.TransactionManager
	repo      domain.UserRepository
}

func NewUserService(txManager domain.TransactionManager, repo domain.UserRepository) *UserService {
	return &UserService{
		txManager: txManager,
		repo:      repo,
	}
}

func (s *UserService) Get(ctx context.Context, userID int) (*domain.User, error) {
	user, err := s.repo.GetByID(ctx, userID, false)

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
	var result_user *domain.User

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.repo.GetByID(ctx, userID, true)
		if err != nil {
			return err
		}

		if user.Status != domain.StatusActive {
			return domain.ErrUserInvalidStatusTransition
		}

		user.Status = domain.StatusBlocked

		err = s.repo.Update(ctx, user)
		if err != nil {
			return err
		}

		result_user = user
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("UserService.Block: %w", err)
	}

	return result_user, nil
}

func (s *UserService) Unblock(ctx context.Context, userID int) (*domain.User, error) {
	var result_user *domain.User

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.repo.GetByID(ctx, userID, true)
		if err != nil {
			return err
		}

		if user.Status != domain.StatusBlocked {
			return domain.ErrUserInvalidStatusTransition
		}

		user.Status = domain.StatusActive

		err = s.repo.Update(ctx, user)
		if err != nil {
			return err
		}

		result_user = user
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("UserService.Unblock: %w", err)
	}

	return result_user, nil
}

func (s *UserService) SetBalance(ctx context.Context, userID, amount int) (*domain.User, error) {
	var result_user *domain.User

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		user, err := s.repo.GetByID(ctx, userID, true)
		if err != nil {
			return err
		}

		user.Balance = amount

		err = s.repo.Update(ctx, user)
		if err != nil {
			return err
		}

		result_user = user
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("UserService.SetBalance: %w", err)
	}

	return result_user, nil
}
