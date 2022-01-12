package eventstream

import (
	"context"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

type EventStream struct {
	Network string
	Address string
}

func New(network string, address string) *EventStream {
	return &EventStream{
		Network: network,
		Address: address,
	}
}

func (e *EventStream) Produce(topic string, message []byte) error {
	// to produce messages
	partition := 0

	conn, err := kafka.DialLeader(context.Background(), e.Network, e.Address, topic, partition)
	if err != nil {
		return fmt.Errorf("failed to dial leader: %w", err)
	}
	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Value: message},
	)
	if err != nil {
		return fmt.Errorf("failed to write messages: %w", err)
	}

	if err := conn.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}
	return nil
}
