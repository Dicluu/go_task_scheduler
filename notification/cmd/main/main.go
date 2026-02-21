package main

import (
	"notification/internal/config"
	"notification/internal/lib/logger/logger"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("log initialized")
}
