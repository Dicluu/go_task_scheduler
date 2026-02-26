package kafka

import (
	"errors"
	"fmt"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	flushTimeout = 5000
)

type Handler interface {
	PrepareMessage(message any) ([]byte, error)
}

var ErrUnknownType = errors.New("unknown event type")
var ErrWrongTypeProvided = errors.New("wrong type of object provided")

type Producer struct {
	producer *kafka.Producer
}

func NewProducer(addresses []string) (*Producer, error) {
	const op = "pkg.kafka.producer.New"

	conf := &kafka.ConfigMap{
		"bootstrap.servers": strings.Join(addresses, ","),
	}

	p, err := kafka.NewProducer(conf)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Producer{
		producer: p,
	}, nil
}

func (p *Producer) Produce(h Handler, message any, topic string) error {
	const op = "pkg.kafka.producer.Produce"

	prepared, err := h.PrepareMessage(message)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	kafkaMessage := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: kafka.PartitionAny,
		},
		Value: prepared,
		Key:   nil,
	}

	kafkaChan := make(chan kafka.Event)
	err = p.producer.Produce(kafkaMessage, kafkaChan)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	e := <-kafkaChan
	switch ev := e.(type) {
	case *kafka.Message:
		return nil
	case kafka.Error:
		return fmt.Errorf("%s: %w", op, ev)
	default:
		return fmt.Errorf("%s: %w", op, ErrUnknownType)
	}
}

func (p *Producer) Close() {
	p.producer.Flush(flushTimeout)
	p.producer.Close()
}
