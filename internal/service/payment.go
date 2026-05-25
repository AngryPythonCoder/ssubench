package service

import (
	"context"
	"fmt"
	"ssubench/internal/domain"
)

type PaymentService struct {
	repo domain.PaymentRepository
}

func NewPaymentService(repo domain.PaymentRepository) *PaymentService {
	return &PaymentService{repo: repo}
}

func (s *PaymentService) Get(ctx context.Context, paymentID int) (*domain.Payment, error) {
	payment, err := s.repo.GetByID(ctx, paymentID, false)

	if err != nil {
		return nil, fmt.Errorf("PaymentService.Get: %w", err)
	}

	return payment, nil
}

func (s *PaymentService) List(ctx context.Context, limit, offset int) ([]domain.Payment, error) {
	payments, err := s.repo.List(ctx, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("PaymentService.List: %w", err)
	}

	return payments, nil
}
