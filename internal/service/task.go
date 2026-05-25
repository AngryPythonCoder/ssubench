package service

import (
	"context"
	"fmt"
	"ssubench/internal/domain"
)

type TaskService struct {
	txManager   domain.TransactionManager
	repo        domain.TaskRepository
	userRepo    domain.UserRepository
	paymentRepo domain.PaymentRepository
}

func NewTaskService(
	txManager domain.TransactionManager,
	repo domain.TaskRepository,
	userRepo domain.UserRepository,
	paymentRepo domain.PaymentRepository,
) *TaskService {
	return &TaskService{
		txManager:   txManager,
		repo:        repo,
		userRepo:    userRepo,
		paymentRepo: paymentRepo,
	}
}

func (s *TaskService) Create(ctx context.Context, title, description string, reward int, customerID int) (*domain.Task, error) {
	task := &domain.Task{
		Title:       title,
		Description: description,
		Reward:      reward,
		Status:      domain.StatusPublished,
		CustomerID:  customerID,
	}

	err := s.repo.Create(ctx, task)

	if err != nil {
		return nil, fmt.Errorf("TaskService.Create: %w", err)
	}

	return task, nil
}

func (s *TaskService) Get(ctx context.Context, taskID int) (*domain.Task, error) {
	task, err := s.repo.GetByID(ctx, taskID, false)

	if err != nil {
		return nil, fmt.Errorf("TaskService.Get: %w", err)
	}

	return task, nil
}

func (s *TaskService) List(ctx context.Context, limit, offset int) ([]domain.Task, error) {
	tasks, err := s.repo.List(ctx, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("TaskService.List: %w", err)
	}

	return tasks, nil
}

func (s *TaskService) CreateBid(ctx context.Context, taskID, performerID int, text string) (*domain.Bid, error) {
	bid := &domain.Bid{
		TaskID:      taskID,
		PerformerID: performerID,
		Text:        text,
	}

	err := s.repo.CreateBid(ctx, bid)

	if err != nil {
		return nil, fmt.Errorf("TaskService.CreateBid: %w", err)
	}

	return bid, nil
}

func (s *TaskService) GetBid(ctx context.Context, taskID, bidID int) (*domain.Bid, error) {
	_, err := s.repo.GetByID(ctx, taskID, false)

	if err != nil {
		return nil, fmt.Errorf("TaskService.GetBid: %w", err)
	}

	bid, err := s.repo.GetBidByID(ctx, taskID, bidID, false)

	if err != nil {
		return nil, fmt.Errorf("TaskService.GetBid: %w", err)
	}

	return bid, nil
}

func (s *TaskService) ListBids(ctx context.Context, taskID, limit, offset int) ([]domain.Bid, error) {
	_, err := s.repo.GetByID(ctx, taskID, false)

	if err != nil {
		return nil, fmt.Errorf("TaskService.ListBids: %w", err)
	}

	bids, err := s.repo.ListBids(ctx, taskID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("TaskService.ListBids: %w", err)
	}

	return bids, nil
}

func (s *TaskService) AcceptBid(ctx context.Context, taskID, customerID, bidID int) (*domain.Task, error) {
	var result_task *domain.Task

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := s.repo.GetByID(ctx, taskID, true)
		if err != nil {
			return err
		}

		if task.CustomerID != customerID {
			return domain.ErrNotAuthorized
		}

		if task.Status != domain.StatusPublished || task.PerformerID != nil {
			return domain.ErrTaskInvalidStatusTransition
		}

		bid, err := s.repo.GetBidByID(ctx, taskID, bidID, false)
		if err != nil {
			return err
		}

		performerID := bid.PerformerID

		task.Status = domain.StatusInProgress
		task.PerformerID = &performerID

		err = s.repo.Update(ctx, task)
		if err != nil {
			return err
		}

		result_task = task
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("TaskService.AcceptBid: %w", err)
	}

	return result_task, nil
}

func (s *TaskService) Finish(ctx context.Context, taskID, performerID int) (*domain.Task, error) {
	var result_task *domain.Task

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := s.repo.GetByID(ctx, taskID, true)
		if err != nil {
			return err
		}

		if task.PerformerID == nil || *task.PerformerID != performerID {
			return domain.ErrNotAuthorized
		}

		if task.Status != domain.StatusInProgress {
			return domain.ErrTaskInvalidStatusTransition
		}

		task.Status = domain.StatusDone

		err = s.repo.Update(ctx, task)
		if err != nil {
			return err
		}

		result_task = task
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("TaskService.Finish: %w", err)
	}

	return result_task, nil
}

func (s *TaskService) Complete(ctx context.Context, taskID, customerID int) (*domain.Task, error) {
	var result_task *domain.Task

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := s.repo.GetByID(ctx, taskID, true)
		if err != nil {
			return err
		}

		if task.CustomerID != customerID {
			return domain.ErrNotAuthorized
		}

		if task.Status != domain.StatusDone || task.PerformerID == nil {
			return domain.ErrTaskInvalidStatusTransition
		}

		customer, err := s.userRepo.GetByID(ctx, customerID, true)
		if err != nil {
			return err
		}

		if customer.Balance < task.Reward {
			return domain.ErrInsufficientFunds
		}

		performer, err := s.userRepo.GetByID(ctx, *task.PerformerID, true)
		if err != nil {
			return err
		}

		task.Status = domain.StatusCompleted
		customer.Balance -= task.Reward
		performer.Balance += task.Reward

		err = s.repo.Update(ctx, task)
		if err != nil {
			return err
		}

		err = s.userRepo.Update(ctx, customer)
		if err != nil {
			return err
		}

		err = s.userRepo.Update(ctx, performer)
		if err != nil {
			return err
		}

		payment := &domain.Payment{
			TaskID:      task.ID,
			CustomerID:  task.CustomerID,
			PerformerID: *task.PerformerID,
			Amount:      task.Reward,
		}

		err = s.paymentRepo.Create(ctx, payment)
		if err != nil {
			return err
		}

		result_task = task
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("TaskService.Complete: %w", err)
	}

	return result_task, nil
}

func (s *TaskService) Cancel(ctx context.Context, taskID, customerID int) (*domain.Task, error) {
	var result_task *domain.Task

	err := s.txManager.WithinTransaction(ctx, func(txCtx context.Context) error {
		task, err := s.repo.GetByID(ctx, taskID, true)
		if err != nil {
			return err
		}

		if task.CustomerID != customerID {
			return domain.ErrNotAuthorized
		}

		if task.Status != domain.StatusPublished && task.Status != domain.StatusInProgress {
			return domain.ErrTaskInvalidStatusTransition
		}

		task.Status = domain.StatusCanceled

		err = s.repo.Update(ctx, task)
		if err != nil {
			return err
		}

		result_task = task
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("TaskService.Cancel: %w", err)
	}

	return result_task, nil
}
