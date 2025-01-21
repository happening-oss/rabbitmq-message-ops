package rabbitmq

import (
	"fmt"

	"github.com/rabbitmq/amqp091-go"
)

type Consumer struct {
	connection *amqp091.Connection
	channel    *amqp091.Channel
	args       map[string]interface{}
}

// NewSimpleConsumer creates a new simple consumer.
func NewSimpleConsumer(endpoint string, offset ...string) (*Consumer, error) {
	conn, err := amqp091.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to open a channel: %w", err)
	}
	err = ch.Qos(1, 0, false)
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to set QoS: %w", err)
	}

	var args map[string]interface{}
	if len(offset) > 0 {
		args = map[string]interface{}{"x-stream-offset": offset[0]}
	}

	return &Consumer{
		connection: conn,
		channel:    ch,
		args:       args,
	}, nil
}

// Consume starts consuming messages from the queue.
func (s *Consumer) Consume(queue string) (<-chan amqp091.Delivery, error) {
	msgs, err := s.channel.Consume(
		queue,  // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		s.args, // args
	)
	if err != nil {
		return nil, fmt.Errorf("consumer: failed to register a consumer: %w", err)
	}

	return msgs, nil
}

// Close closes the consumer Connection and channel.
func (s *Consumer) Close() error {
	if err := s.channel.Close(); err != nil {
		return fmt.Errorf("consumer: failed to close channel: %w", err)
	}
	if err := s.connection.Close(); err != nil {
		return fmt.Errorf("consumer: failed to close Connection: %w", err)
	}
	return nil
}
