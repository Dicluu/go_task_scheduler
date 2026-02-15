package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"task/internal/application/usecase/deletetask"
	"task/internal/application/usecase/fetchtask"
	"task/internal/application/usecase/fetchtasks"
	"task/internal/application/usecase/savetask"
	"task/internal/application/usecase/updatetask"
	"task/internal/config"
	"task/internal/http-server/handlers/task/create"
	"task/internal/http-server/handlers/task/index"
	"task/internal/http-server/handlers/task/remove"
	"task/internal/http-server/handlers/task/show"
	"task/internal/http-server/handlers/task/update"
	"task/internal/lib/logger/logger"
	"task/internal/lib/logger/sl"
	"task/internal/middlewares/auth"
	"task/internal/storage/sqlite"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		return
	}

	log.Info("storage initialized")

	r := chi.NewRouter()
	r.Use(middleware.RequestID, middleware.Recoverer, middleware.Logger, auth.Middleware(log, cfg.Secret))

	r.Post("/tasks", create.New(log, savetask.New(storage)))
	r.Get("/tasks/{task}", show.New(log, fetchtask.New(storage)))
	r.Get("/tasks", index.New(log, fetchtasks.New(storage)))
	r.Put("/tasks/{task}", update.New(log, updatetask.New(storage)))
	r.Delete("/tasks/{task}", remove.New(log, deletetask.New(storage)))

	log.Info("router initialized")

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

	log.Info("server started", slog.String("address", cfg.HttpServer.Address))

	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		log.Error("server shutdown failed", sl.Err(err))
	}

	log.Info("server stopped gracefully")
}
