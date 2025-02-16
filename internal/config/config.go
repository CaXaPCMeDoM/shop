package config

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"

	"avito-tech-winter-2025/pkg/rand"
)

type Config struct {
	Jwt          JWT
	Storage      Storage `yaml:"storage"`
	HTTP         HTTP    `yaml:"http"`
	PasswordSalt string  `env:"PASSWORD_SALT" env-required:"true"`
}

type HTTP struct {
	Port     int      `env:"SERVER_PORT" yaml:"port" env-required:"true"`
	Timeouts Timeouts `yaml:"timeouts"`
}

type Timeouts struct {
	ReadHeader time.Duration `yaml:"read_header" env-default:"5s"`
	Read       time.Duration `yaml:"read" env-default:"10s"`
	Write      time.Duration `yaml:"write" env-default:"10s"`
	Idle       time.Duration `yaml:"idle" env-default:"30s"`
}

type Storage struct {
	Postgres Postgres `yaml:"postgres"`
}

type Postgres struct {
	Host     string     `env:"DATABASE_HOST" env-required:"true"`
	Port     int        `env:"DATABASE_PORT" env-required:"true"`
	User     string     `env:"DATABASE_USER" env-required:"true"`
	Password string     `env:"DATABASE_PASSWORD" env-required:"true"`
	DBName   string     `env:"DATABASE_NAME" env-required:"true"`
	SSLMode  string     `env:"SSL_MODE" env-default:"disable"`
	Pool     PoolConfig `env:"pool"`
}

type PoolConfig struct {
	MaxOpenConns    int           `yaml:"max_open_conns" env-default:"100"`
	MaxIdleConns    int           `yaml:"max_idle_conns" env-default:"10"`
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time" env-default:"1m"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime" env-default:"10m"`
}

type JWT struct {
	SecretKey string        `env:"JWT_SECRET_KEY" env-required:"true"`
	TokenTTL  time.Duration `yaml:"token_ttl" env-default:"1h"`
}

const (
	defaultSaltLength = 10
)

func MustLoad() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("no .env file found")
	}
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH is not set")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exists: %s", configPath)
	}

	var cfg Config

	cfg.PasswordSalt = rand.MustStr(defaultSaltLength)

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("Cannot read config: %s", err)
	}

	return &cfg
}

func (p *Postgres) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode,
	)
}
