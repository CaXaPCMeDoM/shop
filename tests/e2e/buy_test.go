package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (s *APITestSuite) TestBuyItem() {
	user := s.createUser("buyer_user", "pass")
	itemName := "t-shirt"
	itemPrice := 80

	s.T().Run("Successful purchase", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/buy/"+itemName, http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var coins int
		err := s.db.DB.QueryRow(
			"SELECT coins FROM users WHERE username = $1",
			user.Username,
		).Scan(&coins)
		assert.NoError(t, err)
		assert.Equal(t, 1000-itemPrice, coins)

		var inventoryCount int
		err = s.db.DB.QueryRow(
			"SELECT COUNT(*) FROM inventory WHERE user_id = (SELECT id FROM users WHERE username = $1)",
			user.Username,
		).Scan(&inventoryCount)
		assert.NoError(t, err)
		assert.Equal(t, 1, inventoryCount)
	})

	s.T().Run("Insufficient funds", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/buy/expensive_item", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	s.T().Run("Invalid item", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/buy/invalid_item", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
