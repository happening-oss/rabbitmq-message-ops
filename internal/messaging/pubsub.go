package messaging

import "github.com/rabbitmq/amqp091-go"

//go:generate mockery --case underscore --name "Publisher|Consumer" --output ./mocks

type Publisher interface {
	Publish(topic string, msg amqp091.Publishing) error
	Close() error
}

type Consumer interface {
	Consume(queue string) (<-chan amqp091.Delivery, error)
	Close() error
}
