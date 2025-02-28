package main

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
)

func (s *APITestSuite) TestSendCoin() {
	sender := s.createUser("sender_user_1")
	receiver := s.createUser("receiver_user_1")

	var initialCoins int
	err := s.db.DB.QueryRow(
		"SELECT coins FROM users WHERE username = $1",
		sender.Username,
	).Scan(&initialCoins)
	s.Require().NoError(err)

	_, err = s.db.DB.Exec(
		"UPDATE users SET coins = $1 WHERE username = $2",
		100,
		sender.Username,
	)
	s.Require().NoError(err)

	s.Run("Successful transfer", func() {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": 50}`, receiver.Username)
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+sender.Token)

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		s.Require().Equal(s.T(), http.StatusOK, w.Code)

		var senderCoins, receiverCoins int
		err = s.db.DB.QueryRow(
			"SELECT coins FROM users WHERE username = $1",
			sender.Username,
		).Scan(&senderCoins)
		s.Require().NoError(err)

		err = s.db.DB.QueryRow(
			"SELECT coins FROM users WHERE username = $1",
			receiver.Username,
		).Scan(&receiverCoins)
		s.Require().NoError(err)

		s.Require().Equal(s.T(), 50, senderCoins)
		s.Require().Equal(s.T(), 1050, receiverCoins)
	})
}

func (s *APITestSuite) TestSendCoinEdgeCases() {
	sender := s.createUser("sender_user2")
	receiver := s.createUser("receiver_user2")

	s.Run("Send all coins", func() {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": 1000}`, receiver.Username)
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+sender.Token)

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		s.Require().Equal(s.T(), http.StatusOK, w.Code)

		var senderCoins, receiverCoins int
		err := s.db.DB.QueryRow("SELECT coins FROM users WHERE username = $1", sender.Username).Scan(&senderCoins)
		if err != nil {
			return
		}
		err = s.db.DB.QueryRow("SELECT coins FROM users WHERE username = $1", receiver.Username).Scan(&receiverCoins)
		if err != nil {
			return
		}

		s.Require().Equal(s.T(), 0, senderCoins)
		s.Require().Equal(s.T(), 2000, receiverCoins)
	})

	s.Run("Send to self", func() {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": 100}`, sender.Username)
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer "+sender.Token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		s.Require().Equal(s.T(), http.StatusBadRequest, w.Code)
	})

	s.Run("Invalid amount", func() {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": -50}`, receiver.Username)
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer "+sender.Token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		s.Require().Equal(s.T(), http.StatusBadRequest, w.Code)
	})
}
