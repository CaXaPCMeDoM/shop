package postgres

import (
	"avito-tech-winter-2025/internal/storage"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"
)

func (r *Storage) TransferCoins(ctx context.Context, fromUserID int, toUsername string, amount int) error {
	const maxRetryAttemptsForTransaction = 3
	const retryDelay = 100 * time.Millisecond

	rollbackWithLog := func(tx *sql.Tx, originalErr error) error {
		errRollback := tx.Rollback()
		if errRollback != nil {
			if !errors.Is(errRollback, sql.ErrTxDone) {
				slog.ErrorContext(ctx, "transaction rollback failed",
					"rollback_error", errRollback,
					"original_error", originalErr,
				)
				log.Printf("[ERROR] rollback failed: %v (original error: %v)", errRollback, originalErr)
			}
		}
		return originalErr
	}

	for attempt := 1; attempt <= maxRetryAttemptsForTransaction; attempt++ {
		tx, err := r.DB.BeginTx(ctx, &sql.TxOptions{Isolation: sql.LevelRepeatableRead})
		if err != nil {
			if isRetryableError(err) {
				time.Sleep(retryDelay)
				continue
			}
			return fmt.Errorf("begin transaction failed: %w", err)
		}

		var toUserID int
		err = tx.QueryRowContext(ctx, "SELECT id FROM users WHERE username = $1", toUsername).Scan(&toUserID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				return rollbackWithLog(tx, storage.ErrUserNotFound)
			}
			return rollbackWithLog(tx, fmt.Errorf("get receiver ID failed: %w", err))
		}

		result, err := tx.ExecContext(ctx, `
            UPDATE users 
            SET coins = CASE 
                WHEN id = $1 AND coins >= $3 THEN coins - $3
                WHEN id = $2 THEN coins + $3 
                ELSE coins 
            END
            WHERE id IN ($1, $2)`, fromUserID, toUserID, amount)
		if err != nil {
			return rollbackWithLog(tx, fmt.Errorf("update balances failed: %w", err))
		}

		rowsAffected, _ := result.RowsAffected()
		if rowsAffected != 2 {
			return rollbackWithLog(tx, storage.ErrInsufficientFunds)
		}

		_, err = tx.ExecContext(ctx, `
            INSERT INTO transactions (from_user_id, to_user_id, amount, created_at) 
            VALUES ($1, $2, $3, NOW())`, fromUserID, toUserID, amount)
		if err != nil {
			return rollbackWithLog(tx, fmt.Errorf("create transaction record failed: %w", err))
		}

		if err = tx.Commit(); err != nil {
			if isRetryableError(err) {
				err = rollbackWithLog(tx, err)
				if err != nil {
					return fmt.Errorf("%w", err)
				}
				time.Sleep(retryDelay)
				continue
			}
			return rollbackWithLog(tx, fmt.Errorf("commit transaction failed: %w", err))
		}

		return nil
	}

	return fmt.Errorf("transaction failed after %d attempts", maxRetryAttemptsForTransaction)
}
