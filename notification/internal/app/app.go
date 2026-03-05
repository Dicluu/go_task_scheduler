package app

import (
	"log/slog"
	seadapter "notification/internal/adapters/sendemail"
	grpcapp "notification/internal/app/grpc"
	"notification/internal/application/usecase/sendemail"
	"notification/internal/config"
	"notification/internal/storage/sqlite"
)

type App struct {
	GRPCServer *grpcapp.App
}

func New(log *slog.Logger, grpcPort int, storagePath string, cfg config.SmtpServer) *App {
	storage, err := sqlite.New(storagePath)
	if err != nil {
		panic(err)
	}

	notifyService := sendemail.New(storage, seadapter.New(cfg), log)

	grpcApp := grpcapp.New(log, grpcPort, notifyService)

	return &App{
		GRPCServer: grpcApp,
	}
}
