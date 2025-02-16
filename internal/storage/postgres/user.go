package postgres

import (
	"avito-tech-winter-2025/internal/storage"
	"database/sql"
	"errors"

	_ "github.com/lib/pq"
)

func (r *Storage) GetByUsername(username string) (*storage.User, error) {
	var user storage.User
	err := r.DB.QueryRow("SELECT id, username, password_hash, coins FROM users WHERE username = $1", username).Scan(
		&user.ID, &user.Username, &user.PasswordHash, &user.Coins,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return &user, nil
}

func (r *Storage) Create(username, passwordHash string) (*storage.User, error) {
	var user storage.User
	err := r.DB.QueryRow(
		"INSERT INTO users (username, password_hash) VALUES ($1, $2) RETURNING id, username, password_hash, coins",
		username, passwordHash,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.Coins)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
