package utilsService

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// WithTx выполняет функцию fn внутри транзакции
func WithTx(ctx context.Context, db *pgxpool.Pool, fn func(conn *pgxpool.Conn) error) error {
	conn, err := db.Acquire(ctx)
	if err != nil {
		return fmt.Errorf("failed to acquire connection: %w", err)
	}
	defer conn.Release()

	tx, err := conn.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(conn); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
