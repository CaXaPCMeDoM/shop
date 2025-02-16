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

func TestBuyItem_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Error creating mock: %v", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {

		}
	}(db)

	repo := postgres.Storage{DB: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT price FROM merch WHERE item = \\$1").
		WithArgs("t-shirt").
		WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(80))

	mock.ExpectQuery("UPDATE users SET coins = coins - \\$1 WHERE id = \\$2 AND coins >= \\$1 RETURNING coins").
		WithArgs(80, 1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(920))

	mock.ExpectExec("INSERT INTO inventory .+ VALUES \\(\\$1, \\$2, 1\\)").
		WithArgs(1, "t-shirt").
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()

	err = repo.BuyItem(context.Background(), 1, "t-shirt")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyItem_InsufficientFunds(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {

		}
	}(db)

	repo := postgres.Storage{DB: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT price FROM merch WHERE item = \\$1").
		WithArgs("t-shirt").
		WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(80))
	mock.ExpectQuery("UPDATE users SET coins = coins - \\$1 WHERE id = \\$2 AND coins >= \\$1 RETURNING coins").
		WithArgs(80, 1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}))
	mock.ExpectRollback()

	err = repo.BuyItem(context.Background(), 1, "t-shirt")
	assert.ErrorIs(t, err, storage.ErrInsufficientFunds)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyItem_InventoryUpdateError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {

		}
	}(db)

	repo := postgres.Storage{DB: db}

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT price FROM merch WHERE item = \\$1").
		WithArgs("t-shirt").
		WillReturnRows(sqlmock.NewRows([]string{"price"}).AddRow(80))
	mock.ExpectQuery("UPDATE users SET coins = coins - \\$1 WHERE id = \\$2 AND coins >= \\$1 RETURNING coins").
		WithArgs(80, 1).
		WillReturnRows(sqlmock.NewRows([]string{"coins"}).AddRow(920))
	mock.ExpectExec("INSERT INTO inventory .+").
		WithArgs(1, "t-shirt").
		WillReturnError(errors.New("inventory error"))
	mock.ExpectRollback()

	err = repo.BuyItem(context.Background(), 1, "t-shirt")
	assert.ErrorContains(t, err, "inventory error")
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuyItem_BeginTxError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatal(err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {

		}
	}(db)

	repo := postgres.Storage{DB: db}

	mock.ExpectBegin().WillReturnError(errors.New("tx error"))

	err = repo.BuyItem(context.Background(), 1, "t-shirt")
	assert.ErrorContains(t, err, "tx error")
	assert.NoError(t, mock.ExpectationsWereMet())
}
