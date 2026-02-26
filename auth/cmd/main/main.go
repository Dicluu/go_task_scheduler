package main

import (
	"auth/internal/config"
	"auth/internal/http-server/handlers/user/login"
	"auth/internal/http-server/handlers/user/refresh"
	reg "auth/internal/http-server/handlers/user/register"
	"auth/internal/lib/logger/logger"
	"auth/internal/lib/logger/sl"
	"auth/internal/storage/sqlite"
	"auth/pkg/kafka"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("starting application", slog.String("env", cfg.Env))

	p, err := kafka.NewProducer(cfg.Kafka.Servers)
	if err != nil {
		panic(err)
	}

	log.Info("kafka initialized", slog.String("hosts", strings.Join(cfg.Kafka.Servers, ",")))

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		return
	}

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger)
	
	r.Post("/register", reg.New(log, reg.NewUsecase(storage, log, p)))
	r.Post("/login", login.New(log, login.NewUsecase(log, storage, storage, cfg.Secret, cfg.TokenTTL, cfg.RefreshTokenTTL)))
	r.Post("/refresh", refresh.New(log, refresh.NewUsecase(log, storage, storage, cfg.TokenTTL, cfg.RefreshTokenTTL), cfg.Secret))

	srv := &http.Server{
		Addr:         cfg.HttpServer.Address,
		Handler:      r,
		ReadTimeout:  cfg.HttpServer.Timeout,
		WriteTimeout: cfg.HttpServer.Timeout,
		IdleTimeout:  cfg.HttpServer.IdleTimeout,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				panic(fmt.Errorf("server cannot run: %v", err))
			}
		}
	}()

	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		log.Error("server shutdown failed", sl.Err(err))
	}

	log.Info("server stopped gracefully")
}
