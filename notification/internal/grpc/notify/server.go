package notify

import (
	"context"
	"io"
	"log/slog"
	"notification/internal/domain/models"
	"notification/internal/grpc/dto"
	"notification/internal/lib/logger/sl"
	notifyv1 "notification/pkg/protos/gen/go/notify"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Notifier interface {
	Send(ctx context.Context, notifications []models.Notification) error
}

type serverAPI struct {
	gctx context.Context
	notifyv1.UnimplementedNotifierServer
	notifier Notifier
	log      *slog.Logger
}

func Register(gctx context.Context, gRPC *grpc.Server, notifier Notifier, log *slog.Logger) {
	notifyv1.RegisterNotifierServer(gRPC, &serverAPI{gctx: gctx, notifier: notifier, log: log})
}

// TODO: make notifies idempodent
func (s *serverAPI) Notify(srv notifyv1.Notifier_NotifyServer) error {
	const op = "internal.grpc.server.Notify"
	log := s.log.With(slog.String("op", op))
	ctx := srv.Context()
	done := make(chan struct{})
	defer close(done)

	go func() {
		for {
			req, err := srv.Recv()
			if err != nil {
				log.Error("gRPC stream closed by interrupt")
				return
			}
			if err == io.EOF {
				log.Info("stream closed successfully by client")
				return
			}

			if err := validateItems(req); err != nil {
				log.Warn("missing request items")
			}

			recipients := make([]models.Notification, 0)

			items := req.GetItems()
			for _, i := range items {
				date, err := time.Parse("2006-01-02 15:04:05", i.Date)
				if err != nil {
					log.Warn("failed to parse date", slog.String("date", i.Date), sl.Err(err))

					continue
				}

				transfer := dto.NotifyRequestDTO{
					UserId:   i.UserId,
					TaskName: i.Name,
					Date:     date,
				}

				recipients = append(recipients, *transfer.ToDomain())
			}

			err = s.notifier.Send(ctx, recipients)
			if err != nil {
				log.Error("failed to send notifications", sl.Err(err))

				continue
			}

			log.Info("notifications are sent")
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
			log.Info("client side ctx done")
			done <- struct{}{}
		case <-s.gctx.Done():
			log.Info("server side ctx canceled, RPC stream closes")

			err := srv.SendAndClose(&notifyv1.NotifyResponse{
				Success: false,
			})

			if err != nil {
				log.Error("failed to response client side", sl.Err(err))
			}

			done <- struct{}{}
		}
	}()

	<-done
	return nil
}

// unary RPC implement
/*func (s *serverAPI) Notify(ctx context.Context, req *notifyv1.NotifyRequest) (*notifyv1.NotifyResponse, error) {
	const op = "internal.groc.server.Notify"
	log := s.log.With("op", op)

	if err := validateItems(req); err != nil {
		return nil, err
	}

	recipients := make([]models.Notification, 0)

	items := req.GetItems()
	for _, i := range items {
		date, err := time.Parse("2006-01-02 15:04:05", i.Date)
		if err != nil {
			log.Info("failed to parse date", slog.String("date", i.Date), sl.Err(err))

			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("failed to parse date %s", i.Date))
		}

		transfer := dto.NotifyRequestDTO{
			UserId:   i.UserId,
			TaskName: i.Name,
			Date:     date,
		}

		recipients = append(recipients, *transfer.ToDomain())
	}

	err := s.notifier.Send(ctx, recipients)
	if err != nil {
		log.Error("failed to send notifies", sl.Err(err))

		return nil, status.Error(codes.Internal, "failed to send notifies")
	}

	s.log.Info("notifies are sent")

	return &notifyv1.NotifyResponse{
		Success: true,
	}, nil
}*/

func validateItems(req *notifyv1.NotifyRequest) error {
	if len(req.GetItems()) == 0 {
		return status.Error(codes.InvalidArgument, "items is required")
	}

	return nil
}
