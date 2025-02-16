package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

const shutdownTimeout = 5 * time.Second

func SetupServer(deps *Dependencies) {
	gin.SetMode(gin.ReleaseMode)
	router := SetupRouter(deps)

	serverAddr := ":" + strconv.Itoa(deps.Cfg.HTTP.Port)
	srv := &http.Server{
		Addr:              serverAddr,
		Handler:           router,
		ReadHeaderTimeout: deps.Cfg.HTTP.Timeouts.ReadHeader,
		ReadTimeout:       deps.Cfg.HTTP.Timeouts.Read,
		WriteTimeout:      deps.Cfg.HTTP.Timeouts.Write,
		IdleTimeout:       deps.Cfg.HTTP.Timeouts.Idle,
	}

	go func() {
		slog.Info("starting server", slog.String("address", serverAddr))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			slog.Error("server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	slog.Info("stopping application", slog.String("signal", sign.String()))

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		slog.Error("server shutdown error", slog.String("error", err.Error()))
	}

	slog.Info("app stopped")
}
