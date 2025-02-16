package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func (s *APITestSuite) TestAuth() {
	s.Run("Successful registration and auth", func() {
		reqBody := `{"username": "testuser1", "password": "testpass"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/auth", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		s.server.ServeHTTP(w, req)
		s.Require().Equal(s.T(), http.StatusOK, w.Code)

		var authResp struct{ Token string }
		err := json.Unmarshal(w.Body.Bytes(), &authResp)
		if err != nil {
			return
		}
		s.Require().NotEmpty(s.T(), authResp.Token)
	})

	s.Run("Existing user auth", func() {
		reqBody := `{"username": "testuser1", "password": "testpass"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/auth", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		s.server.ServeHTTP(w, req)
		s.Require().Equal(s.T(), http.StatusOK, w.Code)
	})

	s.Run("Invalid credentials", func() {
		reqBody := `{"username": "testuser1", "password": "wrong"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/auth", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		s.server.ServeHTTP(w, req)
		s.Require().Equal(s.T(), http.StatusUnauthorized, w.Code)
	})
}
