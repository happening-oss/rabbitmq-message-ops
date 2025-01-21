package handlers

import (
	"github.com/rabbitmq/amqp091-go"
)

//go:generate mockery --case underscore --name "MessageHandler" --output ./mocks

type MessageHandler interface {
	Handle(msg amqp091.Delivery) (requeue bool, err error)
}
