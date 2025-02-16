package main

import (
	"avito-tech-winter-2025/internal/config"
	"avito-tech-winter-2025/internal/http/api"
	"avito-tech-winter-2025/internal/http/auth"
	"avito-tech-winter-2025/internal/storage/postgres"
	"avito-tech-winter-2025/pkg/hash"
	"log/slog"
	"os"
)

func main() {
	cfg := config.MustLoad()

	db, err := postgres.New(&cfg.Storage.Postgres)
	if err != nil {
		slog.Error("failed to init storage", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.DB.Close()

	tokenMgr := auth.NewManager(cfg)
	hasher := hash.NewSHA1(cfg.PasswordSalt)

	deps := &api.Dependencies{
		DB:       db,
		TokenMgr: tokenMgr,
		Hasher:   hasher,
		Cfg:      cfg,
	}

	api.SetupServer(deps)
}
