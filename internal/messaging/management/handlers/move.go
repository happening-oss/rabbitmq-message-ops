package handlers

import (
	"github.com/rabbitmq/amqp091-go"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging"
	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/mappers"
)

type MoveHandler struct {
	publisher messaging.Publisher
	destQueue string
}

func NewMoveHandler(publisher messaging.Publisher, destQueue string) *MoveHandler {
	return &MoveHandler{publisher: publisher, destQueue: destQueue}
}

func (h *MoveHandler) Handle(msg amqp091.Delivery) (bool, error) {
	// move/publish message to the destination queue
	err := h.publisher.Publish(h.destQueue, mappers.DeliveryPublishing(msg))
	if err != nil {
		return true, err
	}
	return false, nil
}
