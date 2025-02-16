package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func (s *APITestSuite) TestAuth() {
	s.T().Run("Successful registration and auth", func(t *testing.T) {
		reqBody := `{"username": "testuser1", "password": "testpass"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/auth", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		s.server.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)

		var authResp struct{ Token string }
		err := json.Unmarshal(w.Body.Bytes(), &authResp)
		if err != nil {
			return
		}
		assert.NotEmpty(t, authResp.Token)
	})

	s.T().Run("Existing user auth", func(t *testing.T) {
		reqBody := `{"username": "testuser1", "password": "testpass"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/auth", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		s.server.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	})

	s.T().Run("Invalid credentials", func(t *testing.T) {
		reqBody := `{"username": "testuser1", "password": "wrong"}`
		w := httptest.NewRecorder()
		req, _ := http.NewRequestWithContext(context.Background(), "POST", "/api/auth", bytes.NewBufferString(reqBody))
		req.Header.Set("Content-Type", "application/json")

		s.server.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
}
