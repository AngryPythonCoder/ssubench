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

type PGXTaskRepository struct {
	db *pgxpool.Pool
}

func NewPGXTaskRepository(db *pgxpool.Pool) *PGXTaskRepository {
	return &PGXTaskRepository{db: db}
}

func (r *PGXTaskRepository) getQuerier(ctx context.Context) domain.Querier {
	tx, ok := ctx.Value(domain.TransactionKey).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *PGXTaskRepository) Create(ctx context.Context, task *domain.Task) error {
	q := r.getQuerier(ctx)

	query := `INSERT INTO tasks (title, description, reward, status, customer_id)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, performer_id, created_at`

	err := q.QueryRow(ctx, query,
		task.Title,
		task.Description,
		task.Reward,
		task.Status,
		task.CustomerID,
	).Scan(
		&task.ID,
		&task.PerformerID,
		&task.CreatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("PGXTaskRepository.Create: %w", domain.ErrTaskAlreadyExists)
			}
		}

		return fmt.Errorf("PGXTaskRepository.Create: %w", err)
	}

	return nil
}

func (r *PGXTaskRepository) GetByID(ctx context.Context, taskID int, forUpdate bool) (*domain.Task, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, title, description, reward, status, customer_id, performer_id, created_at
			FROM tasks
			where id = $1`

	if forUpdate {
		query += ` FOR UPDATE`
	}

	var task domain.Task
	err := q.QueryRow(ctx, query, taskID).Scan(
		&task.ID,
		&task.Title,
		&task.Description,
		&task.Reward,
		&task.Status,
		&task.CustomerID,
		&task.PerformerID,
		&task.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("PGXTaskRepository.GetByID: %w", domain.ErrTaskNotFound)
		}

		return nil, fmt.Errorf("PGXTaskRepository.GetByID: %w", err)
	}

	return &task, nil
}

func (r *PGXTaskRepository) List(ctx context.Context, limit, offset int) ([]domain.Task, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, title, description, reward, status, customer_id, performer_id, created_at
			FROM tasks
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

	rows, err := q.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("PGXTaskRepository.List: %w", err)
	}

	tasks := []domain.Task{}

	defer rows.Close()
	for rows.Next() {
		var task domain.Task
		err := rows.Scan(
			&task.ID,
			&task.Title,
			&task.Description,
			&task.Reward,
			&task.Status,
			&task.CustomerID,
			&task.PerformerID,
			&task.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("PGXTaskRepository.List: %w", err)
		}

		tasks = append(tasks, task)
	}

	return tasks, nil
}

func (r *PGXTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	q := r.getQuerier(ctx)

	query := `
			UPDATE tasks
			SET
				status = $1,
				performer_id = $2
			WHERE
				id = $3`

	_, err := q.Exec(ctx, query, task.Status, task.PerformerID, task.ID)
	if err != nil {
		return fmt.Errorf("PGXTaskRepository.Update: %w", err)
	}

	return nil
}

func (r *PGXTaskRepository) CreateBid(ctx context.Context, bid *domain.Bid) error {
	q := r.getQuerier(ctx)

	query := `INSERT INTO bids (task_id, performer_id, text)
			  VALUES ($1, $2, $3) RETURNING id, created_at`

	err := q.QueryRow(ctx, query,
		bid.TaskID,
		bid.PerformerID,
		bid.Text,
	).Scan(&bid.ID, &bid.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23503":
				return fmt.Errorf("PGXTaskRepository.CreateBid: %w", domain.ErrTaskNotFound)
			case "23505":
				return fmt.Errorf("PGXTaskRepository.CreateBid: %w", domain.ErrBidAlreadyExists)
			}
		}

		return fmt.Errorf("PGXTaskRepository.CreateBid: %w", err)
	}

	return nil
}

func (r *PGXTaskRepository) GetBidByID(ctx context.Context, taskID, bidID int, forUpdate bool) (*domain.Bid, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, task_id, performer_id, text, created_at
			FROM bids
			WHERE task_id = $1 AND id = $2`

	if forUpdate {
		query += ` FOR UPDATE`
	}

	var bid domain.Bid
	err := q.QueryRow(ctx, query, taskID, bidID).Scan(
		&bid.ID,
		&bid.TaskID,
		&bid.PerformerID,
		&bid.Text,
		&bid.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("PGXTaskRepository.GetBidByID: %w", domain.ErrBidNotFound)
		}

		return nil, fmt.Errorf("PGXTaskRepository.GetBidByID: %w", err)
	}

	return &bid, nil
}

func (r *PGXTaskRepository) ListBids(ctx context.Context, taskID, limit, offset int) ([]domain.Bid, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, task_id, performer_id, text, created_at
			FROM bids
			WHERE task_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3`

	rows, err := q.Query(ctx, query, taskID, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("PGXTaskRepository.ListBids: %w", err)
	}

	bids := []domain.Bid{}

	defer rows.Close()
	for rows.Next() {
		var bid domain.Bid
		err := rows.Scan(
			&bid.ID,
			&bid.TaskID,
			&bid.PerformerID,
			&bid.Text,
			&bid.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("PGXTaskRepository.ListBids %w", err)
		}

		bids = append(bids, bid)
	}

	return bids, nil
}
