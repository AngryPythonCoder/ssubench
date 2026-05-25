package domain

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, arguments ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, arguments ...any) pgx.Row
}

type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByUsername(ctx context.Context, username string, forUpdate bool) (*User, error)
	GetByID(ctx context.Context, userID int, forUpdate bool) (*User, error)
	List(ctx context.Context, limit, offset int) ([]User, error)
	Update(ctx context.Context, user *User) error
}

type TaskRepository interface {
	Create(ctx context.Context, task *Task) error
	GetByID(ctx context.Context, taskID int, forUpdate bool) (*Task, error)
	List(ctx context.Context, limit, offset int) ([]Task, error)
	Update(ctx context.Context, task *Task) error
	CreateBid(ctx context.Context, bid *Bid) error
	GetBidByID(ctx context.Context, taskID, bidID int, forUpdate bool) (*Bid, error)
	ListBids(ctx context.Context, taskID, limit, offset int) ([]Bid, error)
}

type PaymentRepository interface {
	Create(ctx context.Context, payment *Payment) error
	GetByID(ctx context.Context, paymentID int, forUpdate bool) (*Payment, error)
	List(ctx context.Context, limit, offset int) ([]Payment, error)
}
