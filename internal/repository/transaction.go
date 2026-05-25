package repository

import (
	"context"
	"fmt"
	"ssubench/internal/domain"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PGXTXManager struct {
	pool *pgxpool.Pool
}

func NewPGXTXManager(pool *pgxpool.Pool) *PGXTXManager {
	return &PGXTXManager{pool: pool}
}

func (tm *PGXTXManager) WithinTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("PGXTXManager.WithinTransaction: %w", err)
	}

	txCtx := context.WithValue(ctx, domain.TransactionKey, tx)

	defer func() {
		p := recover()
		if p != nil {
			tx.Rollback(ctx)
			panic(p)
		}
	}()

	err = fn(txCtx)
	if err != nil {
		tx.Rollback(ctx)
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("PGXTXManager.WithinTransaction: %w", err)
	}

	return nil
}
