package main

import (
	updateUc "notification/internal/application/usecase/user/update"
	"notification/internal/config"
	kafkaIn "notification/internal/kafka"
	"notification/internal/kafka/handlers/user/update"
	"notification/internal/lib/logger/logger"
	"notification/internal/lib/logger/sl"
	"notification/internal/storage/sqlite"
	"notification/pkg/kafka"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	log.Info("log initialized")

	l := logger.SetupLogger("local")
	l.Info("logger init")

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		panic(err)
	}

	h := update.New(updateUc.New(storage), l)
	c, err := kafka.NewConsumer(h, l, kafkaIn.UserCreateTopic, "notification", cfg.Kafka.Servers)
	if err != nil {
		panic(err)
	}

	go c.Start()

	sChan := make(chan os.Signal)

	signal.Notify(sChan, syscall.SIGINT, syscall.SIGTERM)

	<-sChan

	err = c.Stop()
	if err != nil {
		log.Error("failed to close consumer", sl.Err(err))
	}

}
