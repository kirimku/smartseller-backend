package database

import (
	"context"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TxFn represents a function that can be executed within a transaction
type TxFn func(ctx context.Context, tx *sqlx.Tx) error

// WithTransaction executes the given function within a database transaction
func WithTransaction(ctx context.Context, db *sqlx.DB, fn TxFn) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p) // re-throw panic after rollback
		}
	}()

	if err := fn(ctx, tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("rollback error: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}
