package main

import (
	"bytes"
	"fmt"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

func (s *APITestSuite) TestSendCoin() {
	sender := s.createUser("sender_user_1", "pass")
	receiver := s.createUser("receiver_user_1", "pass")

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

	s.T().Run("Successful transfer", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": 50}`, receiver.Username)
		req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+sender.Token)

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var senderCoins, receiverCoins int
		err = s.db.DB.QueryRow(
			"SELECT coins FROM users WHERE username = $1",
			sender.Username,
		).Scan(&senderCoins)
		assert.NoError(t, err)

		err = s.db.DB.QueryRow(
			"SELECT coins FROM users WHERE username = $1",
			receiver.Username,
		).Scan(&receiverCoins)
		assert.NoError(t, err)

		assert.Equal(t, 50, senderCoins)
		assert.Equal(t, 1050, receiverCoins)
	})
}

func (s *APITestSuite) TestSendCoinEdgeCases() {
	sender := s.createUser("sender_user2", "pass")
	receiver := s.createUser("receiver_user2", "pass")

	s.T().Run("Send all coins", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": 1000}`, receiver.Username)
		req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+sender.Token)

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var senderCoins, receiverCoins int
		err := s.db.DB.QueryRow("SELECT coins FROM users WHERE username = $1", sender.Username).Scan(&senderCoins)
		if err != nil {
			return
		}
		err = s.db.DB.QueryRow("SELECT coins FROM users WHERE username = $1", receiver.Username).Scan(&receiverCoins)
		if err != nil {
			return
		}

		assert.Equal(t, 0, senderCoins)
		assert.Equal(t, 2000, receiverCoins)
	})

	s.T().Run("Send to self", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": 100}`, sender.Username)
		req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer "+sender.Token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	s.T().Run("Invalid amount", func(t *testing.T) {
		reqBody := fmt.Sprintf(`{"toUser": "%q", "amount": -50}`, receiver.Username)
		req, _ := http.NewRequest("POST", "/api/sendCoin", bytes.NewBufferString(reqBody))
		req.Header.Set("Authorization", "Bearer "+sender.Token)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
