package postgres

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/assert"
)

func TestCreateUser(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer func(db *sql.DB) {
		err = db.Close()
		if err != nil {
			return
		}
	}(db)

	repo := &Storage{DB: db}

	tests := []struct {
		name        string
		username    string
		password    string
		mockClosure func()
		wantErr     bool
	}{
		{
			name:     "successful creation",
			username: "testuser",
			password: "testpass",
			mockClosure: func() {
				mock.ExpectQuery("INSERT INTO users").
					WithArgs("testuser", sqlmock.AnyArg()).
					WillReturnRows(sqlmock.NewRows([]string{"id", "username", "password_hash", "coins"}).
						AddRow(1, "testuser", "hash", 1000))
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockClosure()
			_, err := repo.Create(tt.username, tt.password)
			assert.NoError(t, err)
			assert.NoError(t, mock.ExpectationsWereMet())
		})
	}
}
