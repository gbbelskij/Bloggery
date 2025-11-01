package main

import (
	"context"
	"log/slog"
	"net/http"
	"new_service/internal/config"
	"new_service/internal/handlers/add_post"
	"new_service/internal/handlers/auth"
	getnextposts "new_service/internal/handlers/getPosts"
	"new_service/internal/handlers/registration"
	sl "new_service/internal/lib/logger"
	"new_service/internal/repository/storage"
	jwt_auth "new_service/pkg/auth"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func main() {
	os.Setenv("CONFIG_PATH", "./config/local.yaml")
	cfg := config.MustLoad()

	log := setUpLogger(cfg.Env)
	log.Info("Logger started")

	storage, err := storage.New(cfg.PostgresConnString)
	if err != nil {
		log.Info("Failed to connect to database", sl.Error(err))
		os.Exit(1)
	}
	defer storage.Conn.Close()

	router := gin.Default()

	router.POST("/registration", registration.New(storage, log))
	router.POST("/auth", auth.New(log, cfg, storage))

	protected := router.Group("/protected")
	protected.Use(jwt_auth.JWTAuthMiddleware(cfg.JWTSecret))
	{
		protected.POST("/save_post", add_post.New(log, storage))
		protected.GET("/next_posts", getnextposts.New(log, storage))
	}

	srv := &http.Server{
		Addr:    cfg.Address,
		Handler: router,
	}

	// Канал для ошибок из горутины сервера
	serverErrors := make(chan error, 1)

	go func() {
		log.Info("starting server", slog.String("address", cfg.Address))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("failed to start server", sl.Error(err))
			serverErrors <- err
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	// Ждём либо сигнала, либо ошибки из горутины
	select {
	case sig := <-stop:
		log.Info("received signal", slog.String("signal", sig.String()))
	case err := <-serverErrors:
		log.Error("server error received", sl.Error(err))
		os.Exit(1)
	}

	log.Info("initiating graceful shutdown")

	// КРИТИЧНО: используем контекст с таймаутом
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Error("shutdown error", sl.Error(err))
	}

	log.Info("server stopped gracefully")
}

func setUpLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envDev:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}
	return log
}
