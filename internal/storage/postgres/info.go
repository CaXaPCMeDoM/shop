package postgres

import (
	"avito-tech-winter-2025/internal/storage"
	"context"
	"log/slog"
)

func (r *Storage) GetUserInfo(ctx context.Context, userID int) (*storage.UserInfo, error) {
	info := &storage.UserInfo{
		Inventory: make([]storage.InventoryItem, 0),
		CoinHistory: storage.CoinHistory{
			Received: make([]storage.Transaction, 0),
			Sent:     make([]storage.Transaction, 0),
		},
	}

	if err := r.DB.QueryRowContext(ctx, "SELECT coins FROM users WHERE id = $1", userID).
		Scan(&info.Coins); err != nil {
		return nil, err
	}

	rows, err := r.DB.QueryContext(ctx, "SELECT item, quantity FROM inventory WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "failed to close inventory rows", "error", closeErr)
		}
	}()

	for rows.Next() {
		var item storage.InventoryItem
		if scanErr := rows.Scan(&item.Type, &item.Quantity); scanErr != nil {
			return nil, scanErr
		}
		info.Inventory = append(info.Inventory, item)
	}

	rows, err = r.DB.QueryContext(ctx, `
        SELECT u.username, t.amount, t.created_at 
        FROM transactions t
        JOIN users u ON u.id = t.from_user_id
        WHERE t.to_user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "failed to close received transactions rows", "error", closeErr)
		}
	}()

	for rows.Next() {
		var t storage.Transaction
		if scanErr := rows.Scan(&t.FromUser, &t.Amount, &t.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		info.CoinHistory.Received = append(info.CoinHistory.Received, t)
	}

	rows, err = r.DB.QueryContext(ctx, `
        SELECT u.username, t.amount, t.created_at 
        FROM transactions t
        JOIN users u ON u.id = t.to_user_id
        WHERE t.from_user_id = $1`, userID)
	if err != nil {
		return nil, err
	}
	defer func() {
		if closeErr := rows.Close(); closeErr != nil {
			slog.ErrorContext(ctx, "failed to close sent transactions rows", "error", closeErr)
		}
	}()

	for rows.Next() {
		var t storage.Transaction
		if scanErr := rows.Scan(&t.ToUser, &t.Amount, &t.CreatedAt); scanErr != nil {
			return nil, scanErr
		}
		info.CoinHistory.Sent = append(info.CoinHistory.Sent, t)
	}

	return info, nil
}
