package main

import (
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

	application := app.New(log, cfg.GRPCServer.Port, cfg.StoragePath, cfg.SmtpServer)

	go application.GRPCServer.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	<-stop

	// TODO: add graceful shutdown when rpc stream is alive
	application.GRPCServer.Stop()

	log.Info("application stopped")
}
