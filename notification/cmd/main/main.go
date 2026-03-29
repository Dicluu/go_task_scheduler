package main

import (
	"context"
	"notification/internal/app"
	"notification/internal/config"
	"notification/internal/lib/logger/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("log initialized")

	gctx, cancel := context.WithCancel(context.Background())
	application := app.New(gctx, log, cfg.GRPCServer.Port, cfg.StoragePath, cfg.SmtpServer, cfg.GRPCServer.Address)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	cancel()
	application.GRPCServer.Stop()

	log.Info("application stopped")
}
