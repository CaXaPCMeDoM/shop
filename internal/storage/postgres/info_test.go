package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"avito-tech-winter-2025/internal/storage"
	"avito-tech-winter-2025/internal/storage/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestGetUserInfo_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	repo := postgres.Storage{DB: db}

	userID := 1
	expected := &storage.UserInfo{
		Coins: 1000,
		Inventory: []storage.InventoryItem{
			{Type: "t-shirt", Quantity: 2},
			{Type: "cup", Quantity: 1},
		},
		CoinHistory: storage.CoinHistory{
			Received: []storage.Transaction{
				{
					FromUser:  "user2",
					Amount:    500,
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			},
			Sent: []storage.Transaction{
				{
					ToUser:    "user3",
					Amount:    200,
					CreatedAt: time.Now().Format(time.RFC3339),
				},
			},
		},
	}

	mock.ExpectQuery("SELECT coins FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(expected.Coins))

	mock.ExpectQuery("SELECT item, quantity FROM inventory WHERE user_id = \\$1").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"item", "quantity"}).
			AddRow("t-shirt", 2).
			AddRow("cup", 1))

	mock.ExpectQuery("SELECT u.username, t.amount, t.created_at FROM transactions t").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"username", "amount", "created_at"}).
			AddRow("user2", 500, expected.CoinHistory.Received[0].CreatedAt))

	mock.ExpectQuery("SELECT u.username, t.amount, t.created_at FROM transactions t").
		WithArgs(userID).
		WillReturnRows(sqlmock.NewRows([]string{"username", "amount", "created_at"}).
			AddRow("user3", 200, expected.CoinHistory.Sent[0].CreatedAt))

	result, err := repo.GetUserInfo(context.Background(), userID)

	require.NoError(t, err)
	require.Equal(t, expected, result)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserInfo_UserNotFound(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	repo := postgres.Storage{DB: db}
	userID := 999

	mock.ExpectQuery("SELECT coins FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnError(sql.ErrNoRows)

	result, err := repo.GetUserInfo(context.Background(), userID)

	assert.Nil(t, result)
	require.ErrorIs(t, err, sql.ErrNoRows)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserInfo_QueryError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	repo := postgres.Storage{DB: db}
	userID := 1

	mock.ExpectQuery("SELECT coins FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnError(sql.ErrConnDone)

	result, err := repo.GetUserInfo(context.Background(), userID)

	assert.Nil(t, result)
	require.ErrorIs(t, err, sql.ErrConnDone)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestGetUserInfo_RowsCloseError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	repo := postgres.Storage{DB: db}
	userID := 1

	rows := sqlmock.NewRows([]string{"coins"}).AddRow(1000).CloseError(errors.New("close error"))

	mock.ExpectQuery("SELECT coins FROM users WHERE id = \\$1").
		WithArgs(userID).
		WillReturnRows(rows)

	_, err = repo.GetUserInfo(context.Background(), userID)

	require.ErrorContains(t, err, "close error")
	require.NoError(t, mock.ExpectationsWereMet())
}
