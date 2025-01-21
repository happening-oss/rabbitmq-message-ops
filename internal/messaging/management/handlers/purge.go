package handlers

import (
	"github.com/rabbitmq/amqp091-go"
)

type PurgeHandler struct{}

func NewPurgeHandler() *PurgeHandler {
	return &PurgeHandler{}
}

func (h *PurgeHandler) Handle(_ amqp091.Delivery) (bool, error) {
	return false, nil
}
