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

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) error {
	query := `INSERT INTO users (username, password_hash, role, status, balance)
			  VALUES ($1, $2, $3, $4, $5) RETURNING id, created_at`

	err := r.db.QueryRow(ctx, query,
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
				return fmt.Errorf("UserRepository.Create: %w", domain.ErrUserAlreadyExists)
			}
		}

		return fmt.Errorf("UserRepository.Create: %w", err)
	}

	return nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (*domain.User, error) {
	query := `
			SELECT id, username, password_hash, role, status, balance, created_at 
			FROM users 
			WHERE username = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, username).
		Scan(
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
			return nil, fmt.Errorf("UserRepository.GetByUsername: %w", domain.ErrUserNotFound)
		}

		return nil, fmt.Errorf("UserRepository.GetByUsername: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetByID(ctx context.Context, userID int) (*domain.User, error) {
	query := `
			SELECT id, username, password_hash, role, status, balance, created_at 
			FROM users 
			WHERE id = $1`

	var user domain.User
	err := r.db.QueryRow(ctx, query, userID).
		Scan(
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
			return nil, fmt.Errorf("UserRepository.GetByID: %w", domain.ErrUserNotFound)
		}

		return nil, fmt.Errorf("UserRepository.GetByID: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int) ([]domain.User, error) {
	query := `
			SELECT id, username, password_hash, role, status, balance, created_at 
			FROM users
			ORDER BY created_at DESC
			LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("UserRepository.List: %w", err)
	}

	var users []domain.User

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
			return nil, fmt.Errorf("UserRepository.List: %w", err)
		}

		users = append(users, user)
	}

	return users, nil
}

func (r *UserRepository) ChangeStatus(ctx context.Context, userID int, status domain.UserStatus) (*domain.User, error) {
	query := `
			UPDATE users
			SET status = $1
			WHERE id = $2
			RETURNING id, username, password_hash, role, status, balance, created_at`

	var user domain.User
	err := r.db.QueryRow(ctx, query, status, userID).
		Scan(
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
			return nil, fmt.Errorf("UserRepository.ChangeStatus: %w", domain.ErrUserNotFound)
		}

		return nil, fmt.Errorf("UserRepository.ChangeStatus: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) SetBalance(ctx context.Context, userID, amount int) (*domain.User, error) {
	query := `
			UPDATE users
			SET balance = $1
			WHERE id = $2
			RETURNING id, username, password_hash, role, status, balance, created_at`

	var user domain.User
	err := r.db.QueryRow(ctx, query, amount, userID).
		Scan(
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
			return nil, fmt.Errorf("UserRepository.SetBalance: %w", domain.ErrUserNotFound)
		}

		return nil, fmt.Errorf("UserRepository.SetBalance: %w", err)
	}

	return &user, nil
}
