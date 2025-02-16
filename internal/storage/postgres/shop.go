package postgres

import (
	"avito-tech-winter-2025/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"
)

func (r *Storage) BuyItem(ctx context.Context, userID int, item string) error {
	var lastErr error
	rollbackWithLog := func(tx *sql.Tx, originalErr error) error {
		errRollback := tx.Rollback()
		if errRollback != nil && !errors.Is(errRollback, sql.ErrTxDone) {
			slog.ErrorContext(ctx, "transaction rollback failed",
				"rollback_error", errRollback,
				"original_error", originalErr,
				"userID", userID,
				"item", item,
			)
		}
		return originalErr
	}

	for attempt := 1; attempt <= maxRetryAttemptsForTransaction; attempt++ {
		tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			lastErr = fmt.Errorf("begin transaction failed: %w", err)
			if isRetryableError(err) {
				time.Sleep(retryDelay)
				continue
			}
			return lastErr
		}

		var price int
		err = tx.QueryRowContext(ctx, "SELECT price FROM merch WHERE item = $1", item).Scan(&price)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return rollbackWithLog(tx, storage.ErrInvalidItem)
			}
			return rollbackWithLog(tx, fmt.Errorf("get item price failed: %w", err))
		}

		var newBalance int
		err = tx.QueryRowContext(ctx, `
            UPDATE users 
            SET coins = coins - $1 
            WHERE id = $2 AND coins >= $1
            RETURNING coins`, price, userID).Scan(&newBalance)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return rollbackWithLog(tx, storage.ErrInsufficientFunds)
			}
			return rollbackWithLog(tx, fmt.Errorf("update balance failed: %w", err))
		}

		_, err = tx.ExecContext(ctx, `
            INSERT INTO inventory (user_id, item, quantity)
            VALUES ($1, $2, 1)
            ON CONFLICT (user_id, item) 
            DO UPDATE SET quantity = inventory.quantity + 1`,
			userID, item)
		if err != nil {
			return rollbackWithLog(tx, fmt.Errorf("update inventory failed: %w", err))
		}

		if err = tx.Commit(); err != nil {
			lastErr = fmt.Errorf("commit transaction failed: %w", err)
			if isRetryableError(err) {
				err = rollbackWithLog(tx, err)
				if err != nil {
					return fmt.Errorf("%w", err)
				}
				time.Sleep(retryDelay)
				continue
			}
			return rollbackWithLog(tx, lastErr)
		}

		return nil
	}

	return fmt.Errorf("transaction failed after %d attempts: %w", maxRetryAttemptsForTransaction, lastErr)
}
