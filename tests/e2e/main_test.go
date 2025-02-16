package main

import (
	"avito-tech-winter-2025/internal/config"
	"avito-tech-winter-2025/internal/http/api"
	"avito-tech-winter-2025/internal/http/auth"
	"avito-tech-winter-2025/internal/storage/postgres"
	"avito-tech-winter-2025/pkg/hash"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/suite"

	"github.com/spf13/viper"
)

type TestDBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func LoadTestDBConfig() *TestDBConfig {
	viper.SetDefault("TEST_DB_HOST", "localhost")
	viper.SetDefault("TEST_DB_PORT", 5433)
	viper.SetDefault("TEST_DB_USER", "postgres")
	viper.SetDefault("TEST_DB_PASSWORD", "password")
	viper.SetDefault("TEST_DB_NAME", "shop_test")
	viper.SetDefault("TEST_DB_SSLMODE", "disable")

	viper.AutomaticEnv()

	return &TestDBConfig{
		Host:     viper.GetString("TEST_DB_HOST"),
		Port:     viper.GetInt("TEST_DB_PORT"),
		User:     viper.GetString("TEST_DB_USER"),
		Password: viper.GetString("TEST_DB_PASSWORD"),
		DBName:   viper.GetString("TEST_DB_NAME"),
		SSLMode:  viper.GetString("TEST_DB_SSLMODE"),
	}
}

func (cfg *TestDBConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DBName,
		cfg.SSLMode)
}

type APITestSuite struct {
	suite.Suite
	server   *gin.Engine
	db       *postgres.Storage
	tokenMgr *auth.Manager
	hasher   *hash.SHA1
	cfg      *config.Config
}

func (s *APITestSuite) SetupSuite() {
	testDBConfig := LoadTestDBConfig()

	s.cfg = &config.Config{
		Storage: config.Storage{
			Postgres: config.Postgres{
				Host:     testDBConfig.Host,
				Port:     testDBConfig.Port,
				User:     testDBConfig.User,
				Password: testDBConfig.Password,
				DBName:   testDBConfig.DBName,
				SSLMode:  testDBConfig.SSLMode,
			},
		},
		PasswordSalt: "salt",
		Jwt: config.JWT{
			SecretKey: "secret",
			TokenTTL:  time.Hour,
		},
	}

	db, err := postgres.New(&s.cfg.Storage.Postgres)
	s.Require().NoError(err)
	s.db = db

	s.tokenMgr = auth.NewManager(s.cfg)
	s.hasher = hash.NewSHA1(s.cfg.PasswordSalt)

	deps := &api.Dependencies{
		DB:       db,
		TokenMgr: s.tokenMgr,
		Hasher:   s.hasher,
		Cfg:      s.cfg,
	}
	s.server = api.SetupRouter(deps)
}

func (s *APITestSuite) TearDownSuite() {
	err := s.db.DB.Close()
	if err != nil {
		return
	}
}

func TestAPISuite(t *testing.T) {
	suite.Run(t, new(APITestSuite))
}

func (s *APITestSuite) createUser(username, password string) *TestUser {
	reqBody := fmt.Sprintf(`{"username": "%q", "password": "%q"}`, username, password)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/auth", bytes.NewBufferString(reqBody))
	req.Header.Set("Content-Type", "application/json")
	s.server.ServeHTTP(w, req)

	var resp struct {
		Token string `json:"token"`
	}
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	if err != nil {
		return nil
	}

	return &TestUser{
		Username: username,
		Token:    resp.Token,
	}
}

func (s *APITestSuite) TearDownTest() {
	_, err := s.db.DB.Exec(`
		TRUNCATE TABLE users, inventory, transactions RESTART IDENTITY CASCADE
	`)
	s.Require().NoError(err)
}

type TestUser struct {
	Username string
	Token    string
}
