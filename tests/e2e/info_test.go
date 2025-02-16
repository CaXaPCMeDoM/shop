package main

import (
	"avito-tech-winter-2025/internal/storage"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (s *APITestSuite) TestGetInfo() {
	user := s.createUser("info_user", "pass")

	s.T().Run("Get initial info", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/info", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		var info storage.UserInfo
		err := json.Unmarshal(w.Body.Bytes(), &info)
		assert.NoError(t, err)

		assert.Equal(t, 1000, info.Coins)
		assert.Empty(t, info.Inventory)
		assert.Empty(t, info.CoinHistory.Received)
		assert.Empty(t, info.CoinHistory.Sent)
	})
}
