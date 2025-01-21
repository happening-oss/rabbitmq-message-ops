package rabbitmq

import (
	"context"
	"fmt"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type Publisher struct {
	conn    *amqp091.Connection
	channel *amqp091.Channel
}

// NewSimplePublisher creates a new simple publisher.
func NewSimplePublisher(endpoint string) (*Publisher, error) {
	conn, err := amqp091.Dial(endpoint)
	if err != nil {
		return nil, fmt.Errorf("publisher: failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("publisher: failed to open a channel: %w", err)
	}

	// put channel into confirm mode so that the client can ensure that all publishing have successfully been received by the server
	if err = ch.Confirm(false); err != nil {
		return nil, fmt.Errorf("publisher: channel could not be put into confirm mode: %w", err)
	}

	return &Publisher{
		conn:    conn,
		channel: ch,
	}, nil
}

// Publish sends a message to the RabbitMQ exchange.
func (p *Publisher) Publish(topic string, msg amqp091.Publishing) error {
	confirm, err := p.channel.PublishWithDeferredConfirmWithContext(
		context.Background(),
		amqp091.DefaultExchange,
		topic,
		false,
		false,
		msg,
	)
	if err != nil {
		return fmt.Errorf("publisher: failed to publish a message: %w", err)
	}

	select {
	case <-confirm.Done():
		pSuccess := confirm.Acked()
		if !pSuccess {
			return fmt.Errorf("publisher: negative confirmation received")
		}
	case <-time.After(time.Second * 10):
		return fmt.Errorf("publisher: waiting for publisher confirmation timed out")
	}

	return nil
}

// Close closes the publisher Connection and channel.
func (p *Publisher) Close() error {
	if err := p.channel.Close(); err != nil {
		return fmt.Errorf("publisher: failed to close channel: %w", err)
	}
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("publisher: failed to close Connection: %w", err)
	}
	return nil
}

// region Helpers

// endregion
