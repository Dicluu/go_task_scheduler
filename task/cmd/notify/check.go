package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"task/internal/application/usecase/notifytask"
	"task/internal/config"
	"task/internal/http-server/adapters/task/notify"
	"task/internal/lib/logger/logger"
	"task/internal/lib/logger/sl"
	"task/internal/storage/sqlite"

	"github.com/go-co-op/gocron/v2"
)

func main() {
	cfg := config.MustLoad()

	log := logger.SetupLogger(cfg.Env)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		return
	}

	validateCfg(&cfg.Cron)

	u := notifytask.New(storage, notify.New(log, cfg.Cron.Address), cfg.Cron.Batch, log)

	s, err := gocron.NewScheduler()
	if err != nil {
		panic("failed to create scheduler")
	}

	_, err = s.NewJob(gocron.CronJob("* * * * *", false), gocron.NewTask(
		func() {
			err := u.NotifyTasks(context.Background())
			if err != nil {
				log.Error("failed to send notifications", sl.Err(err))
			}
		}))
	if err != nil {
		log.Error("failed to send notifications", sl.Err(err))
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	s.Start()

	log.Info("cron started")

	<-stop

	err = s.Shutdown()
	if err != nil {
		log.Error("failed to graceful shutdown", sl.Err(err))

		return
	}

	log.Info("cron stopped gracefully")
}

func validateCfg(cfg *config.Cron) {
	if cfg.Address == "" {
		panic("cron not configured: address not provided")
	}
}
