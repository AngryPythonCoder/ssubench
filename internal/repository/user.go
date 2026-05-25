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

type PGXUserRepository struct {
	db *pgxpool.Pool
}

func NewPGXUserRepository(db *pgxpool.Pool) *PGXUserRepository {
	return &PGXUserRepository{db: db}
}

func (r *PGXUserRepository) getQuerier(ctx context.Context) domain.Querier {
	tx, ok := ctx.Value(domain.TransactionKey).(pgx.Tx)
	if ok {
		return tx
	}

	return r.db
}

func (r *PGXUserRepository) Create(ctx context.Context, user *domain.User) error {
	q := r.getQuerier(ctx)

	query := `INSERT INTO users (username, password_hash, role, status, balance)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := q.QueryRow(ctx, query,
		user.Username,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.Balance,
	).Scan(&user.ID, &user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError

		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("PGXUserRepository.Create: %w", domain.ErrUserAlreadyExists)
			}
		}

		return fmt.Errorf("PGXUserRepository.Create: %w", err)
	}

	return nil
}

func (r *PGXUserRepository) GetByUsername(ctx context.Context, username string, forUpdate bool) (*domain.User, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, username, password_hash, role, status, balance, created_at 
			FROM users 
			WHERE username = $1`

	if forUpdate {
		query += ` FOR UPDATE`
	}

	var user domain.User
	err := q.QueryRow(ctx, query, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.Balance,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("PGXUserRepository.GetByUsername: %w", domain.ErrUserNotFound)
		}

		return nil, fmt.Errorf("PGXUserRepository.GetByUsername: %w", err)
	}

	return &user, nil
}

func (r *PGXUserRepository) GetByID(ctx context.Context, userID int, forUpdate bool) (*domain.User, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, username, password_hash, role, status, balance, created_at 
			FROM users 
			WHERE id = $1`

	if forUpdate {
		query += ` FOR UPDATE`
	}

	var user domain.User
	err := q.QueryRow(ctx, query, userID).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.Status,
		&user.Balance,
		&user.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("PGXUserRepository.GetByID: %w", domain.ErrUserNotFound)
		}

		return nil, fmt.Errorf("PGXUserRepository.GetByID: %w", err)
	}

	return &user, nil
}

func (r *PGXUserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	q := r.getQuerier(ctx)

	query := `
			SELECT id, username, password_hash, role, status, balance, created_at 
			FROM users
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

	rows, err := q.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("PGXUserRepository.List: %w", err)
	}

	users := []domain.User{}

	defer rows.Close()
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Username,
			&user.PasswordHash,
			&user.Role,
			&user.Status,
			&user.Balance,
			&user.CreatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("PGXUserRepository.List: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *PGXUserRepository) Update(ctx context.Context, user *domain.User) error {
	q := r.getQuerier(ctx)

	query := `
			UPDATE users
			SET
				status = $1,
				balance = $2
			WHERE
				id = $3`

	_, err := q.Exec(ctx, query, user.Status, user.Balance, user.ID)
	if err != nil {
		return fmt.Errorf("PGXUserRepository.Update: %w", err)
	}

	return nil
}
