package notify

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"task/internal/domain/models"
	"task/internal/http-server/adapters"
	"task/internal/lib/logger/sl"
	notifyv1 "task/pkg/protos/gen/go/notify"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Adapter struct {
	log  *slog.Logger
	addr string
}

var conn notifyv1.Notifier_NotifyClient

func getConnection(addr string) (notifyv1.Notifier_NotifyClient, error) {
	const op = "internal.http-server.adapters.task.Notify"
	var err error

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("error recovered")
			err = fmt.Errorf("%s: %s", op, r)
		}
	}()

	sync.OnceFunc(func() {
		client, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			panic(err)
		}

		nc := notifyv1.NewNotifierClient(client)
		stream, err := nc.Notify(context.Background())
		if err != nil {
			panic(err)
		}

		conn = stream
	})()

	return conn, err
}

func New(log *slog.Logger, addr string) *Adapter {
	return &Adapter{
		log:  log,
		addr: addr,
	}
}

func (a *Adapter) Notify(tasks []models.Task) error {
	const op = "internal.http-server.adapters.task.Notify"
	log := a.log.With(slog.String("op", op))

	var chunk []*notifyv1.Item

	for _, task := range tasks {
		chunk = append(chunk, &notifyv1.Item{
			Name:   task.Name,
			Date:   task.StartsAt.Format("2006-01-02 15:04:05"),
			UserId: task.UserId,
		})
	}

	stream, err := getConnection(a.addr)
	if err != nil {
		log.Error("failed to establish connection", sl.Err(err))

		return fmt.Errorf("%s: %w", op, adapters.ErrFailedConn)
	}

	ctx := stream.Context()
	done := make(chan bool)

	go func() {
		req := notifyv1.NotifyRequest{
			Items: chunk,
		}

		if err := stream.Send(&req); err != nil {
			log.Error("failed to send request", sl.Err(err))
		}

		if err := stream.CloseSend(); err != nil {
			log.Error("failed to close stream", sl.Err(err))
		}
	}()

	go func() {
		<-ctx.Done()
		if err := ctx.Err(); err != nil {
			log.Error("context got an error", sl.Err(err))
		}
		close(done)
	}()

	return nil
}
