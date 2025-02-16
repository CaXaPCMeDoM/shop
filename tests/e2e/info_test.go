package main

import (
	"avito-tech-winter-2025/internal/storage"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (s *APITestSuite) TestGetInfo() {
	user := s.createUser("info_user")

	s.Run("Get initial info", func() {
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/info", http.NoBody)
		req.Header.Set("Authorization", "Bearer "+user.Token)
		s.server.ServeHTTP(w, req)

		s.Require().Equal(http.StatusOK, w.Code)

		var info storage.UserInfo
		err := json.Unmarshal(w.Body.Bytes(), &info)
		s.Require().NoError(err)

		s.Require().Equal(1000, info.Coins)
		s.Require().Empty(info.Inventory)
		s.Require().Empty(info.CoinHistory.Received)
		s.Require().Empty(info.CoinHistory.Sent)
	})
}
