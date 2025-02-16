package postgres_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"avito-tech-winter-2025/internal/storage"
	"avito-tech-winter-2025/internal/storage/postgres"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestTransferCoins_Success(t *testing.T) {
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
	fromUserID := 1
	toUsername := "recipient"
	amount := 100
	toUserID := 2

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs(toUsername).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(toUserID))

	mock.ExpectExec(`UPDATE users SET coins = CASE`).
		WithArgs(fromUserID, toUserID, amount).
		WillReturnResult(sqlmock.NewResult(0, 2))

	mock.ExpectExec(`INSERT INTO transactions`).
		WithArgs(fromUserID, toUserID, amount).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.TransferCoins(context.Background(), fromUserID, toUsername, amount)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferCoins_RecipientNotFound(t *testing.T) {
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
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs("non_existent").
		WillReturnError(sql.ErrNoRows)
	mock.ExpectRollback()

	err = repo.TransferCoins(context.Background(), 1, "non_existent", 100)
	assert.ErrorIs(t, err, storage.ErrUserNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferCoins_InsufficientFunds(t *testing.T) {
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
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs("recipient").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectExec(`UPDATE users SET coins = CASE`).
		WithArgs(1, 2, 100).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectRollback()

	err = repo.TransferCoins(context.Background(), 1, "recipient", 100)
	assert.ErrorIs(t, err, storage.ErrInsufficientFunds)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestTransferCoins_TransactionCommitError(t *testing.T) {
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
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT id FROM users WHERE username = \$1`).
		WithArgs("recipient").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))

	mock.ExpectExec(`UPDATE users SET coins = CASE`).
		WithArgs(1, 2, 100).
		WillReturnResult(sqlmock.NewResult(0, 2))

	mock.ExpectExec(`INSERT INTO transactions`).
		WithArgs(1, 2, 100).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit().WillReturnError(errors.New("commit error"))

	err = repo.TransferCoins(context.Background(), 1, "recipient", 100)
	assert.ErrorContains(t, err, "commit error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
