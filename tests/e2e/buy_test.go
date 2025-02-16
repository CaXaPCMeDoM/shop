package main

import (
	"context"
	"net/http"
	"net/http/httptest"
)

func (s *APITestSuite) TestBuyItem() {
	user := s.createUser("buyer_user")
	itemName := "t-shirt"
	itemPrice := 80

	s.Run("Successful purchase", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/buy/"+itemName, http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		s.Require().Equal(s.T(), http.StatusOK, w.Code)

		var coins int
		err := s.db.DB.QueryRow(
			"SELECT coins FROM users WHERE username = $1",
			user.Username,
		).Scan(&coins)
		s.Require().NoError(err)
		s.Require().Equal(s.T(), 1000-itemPrice, coins)

		var inventoryCount int
		err = s.db.DB.QueryRow(
			"SELECT COUNT(*) FROM inventory WHERE user_id = (SELECT id FROM users WHERE username = $1)",
			user.Username,
		).Scan(&inventoryCount)
		s.Require().NoError(err)
		s.Require().Equal(s.T(), 1, inventoryCount)
	})

	s.Run("Insufficient funds", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/buy/expensive_item", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		s.Require().Equal(s.T(), http.StatusBadRequest, w.Code)
	})

	s.Run("Invalid item", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/buy/invalid_item", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		s.Require().Equal(s.T(), http.StatusBadRequest, w.Code)
	})
}
