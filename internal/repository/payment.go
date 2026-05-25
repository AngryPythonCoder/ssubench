package repository

import (
	"context"
	"errors"
	"fmt"
	"ssubench/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PGXPaymentRepository struct {
	db *pgxpool.Pool
}

func NewPGXPaymentRepository(db *pgxpool.Pool) *PGXPaymentRepository {
	return &PGXPaymentRepository{db: db}
}

func (r *PGXPaymentRepository) getQuerier(ctx context.Context) domain.Querier {
	tx, ok := ctx.Value(domain.TransactionKey).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *PGXPaymentRepository) Create(ctx context.Context, payment *domain.Payment) error {
	q := r.getQuerier(ctx)

	query := `INSERT INTO payments (task_id, customer_id, performer_id, amount)
			  VALUES ($1, $2, $3, $4) RETURNING id, created_at`

	err := q.QueryRow(ctx, query,
		payment.TaskID,
		payment.CustomerID,
		payment.PerformerID,
		payment.Amount,
	).Scan(&payment.ID, &payment.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("PGXPaymentRepository.Create: %w", domain.ErrPaymentAlreadyExists)
			}
		}

		return fmt.Errorf("PGXPaymentRepository.Create: %w", err)
	}

	return nil
}

func (r *PGXPaymentRepository) GetByID(ctx context.Context, paymentID int, forUpdate bool) (*domain.Payment, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, task_id, customer_id, performer_id, amount, created_at
			FROM payments
			WHERE id = $1`

	if forUpdate {
		query += ` FOR UPDATE`
	}

	var payment domain.Payment
	err := q.QueryRow(ctx, query, paymentID).Scan(
		&payment.ID,
		&payment.TaskID,
		&payment.CustomerID,
		&payment.PerformerID,
		&payment.Amount,
		&payment.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("PGXPaymentRepository.GetByID: %w", domain.ErrPaymentNotFound)
		}

		return nil, fmt.Errorf("PGXPaymentRepository.GetByID: %w", err)
	}

	return &payment, nil
}

func (r *PGXPaymentRepository) List(ctx context.Context, limit, offset int) ([]domain.Payment, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, task_id, customer_id, performer_id, amount, created_at
			FROM payments
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

	rows, err := q.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("PGXPaymentRepository.List: %w", err)
	}

	payments := []domain.Payment{}

	defer rows.Close()
	for rows.Next() {
		var payment domain.Payment
		err := rows.Scan(
			&payment.ID,
			&payment.TaskID,
			&payment.CustomerID,
			&payment.PerformerID,
			&payment.Amount,
			&payment.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("PGXPaymentRepository.List %w", err)
		}

		payments = append(payments, payment)
	}

	return payments, nil
}
