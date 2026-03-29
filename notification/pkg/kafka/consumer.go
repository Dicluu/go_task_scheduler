package kafka

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"notification/internal/lib/logger/sl"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	sessionTimeout   = 7000
	reconnectBackoff = 5000
)

type Handler interface {
	HandleMessage(ctx context.Context, message []byte) error
}

type Consumer struct {
	consumer *kafka.Consumer
	handler  Handler
	stop     bool
	log      *slog.Logger
}

func NewConsumer(h Handler, l *slog.Logger, topic, cGroup string, addresses []string) (*Consumer, error) {
	const op = "pkg.kafka.consumer.NewConsumer"

	cfg := &kafka.ConfigMap{
		"bootstrap.servers":        strings.Join(addresses, ","),
		"group.id":                 cGroup,
		"session.timeout.ms":       sessionTimeout,
		"enable.auto.offset.store": false,
		"enable.auto.commit":       true,
		"auto.commit.interval.ms":  5000,
		"reconnect.backoff.ms":     reconnectBackoff,
		"reconnect.backoff.max.ms": 30000,
		"auto.offset.reset":        "earliest",
	}

	c, err := kafka.NewConsumer(cfg)
	if err != nil {
		return &Consumer{}, fmt.Errorf("%s: %w", op, err)
	}

	err = c.Subscribe(topic, nil)
	if err != nil {
		return &Consumer{}, fmt.Errorf("%s: %w", op, err)
	}

	return &Consumer{
		consumer: c,
		handler:  h,
		log:      l,
	}, nil
}

func (c *Consumer) Start() {
	const op = "pkg.kafka.consumer.Start"
	log := c.log.With(slog.String("op", op))
	for {
		if c.stop {
			break
		}

		msg, err := c.consumer.ReadMessage(time.Second)
		if err != nil {
			var e kafka.Error
			errors.As(err, &e)

			if e.Code() == kafka.ErrAllBrokersDown {
				log.Error("failed to read message from kafka, reconnect...")

				// force sleep to prevent read request by timeout before reconnect backoff
				time.Sleep(reconnectBackoff * time.Millisecond)

				continue
			}

			if e.Code() == kafka.ErrTimedOut {
				continue
			}

			log.Error("failed to read message", sl.Err(err))
			continue
		}

		if msg == nil {
			continue
		}

		log.Info("message received", slog.Any("msg", msg))
		if err = c.handler.HandleMessage(context.TODO(), msg.Value); err != nil {
			log.Error("failed to handle message", sl.Err(err))
			continue
		}

		// local commit offset
		if _, err = c.consumer.StoreMessage(msg); err != nil {
			log.Error("failed to commit offset", sl.Err(err))
			continue
		}
	}
}

func (c *Consumer) Stop() error {
	c.stop = true

	if _, err := c.consumer.Commit(); err != nil {
		return err
	}

	return c.consumer.Close()
}
