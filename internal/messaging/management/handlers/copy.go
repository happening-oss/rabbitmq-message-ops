package handlers

import (
	"github.com/rabbitmq/amqp091-go"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging"
	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/mappers"
)

type CopyHandler struct {
	publisher messaging.Publisher
	destQueue string
}

func NewCopyHandler(publisher messaging.Publisher, destQueue string) *CopyHandler {
	return &CopyHandler{publisher: publisher, destQueue: destQueue}
}

func (h *CopyHandler) Handle(msg amqp091.Delivery) (bool, error) {
	// copy/publish message to the destination queue
	return true, h.publisher.Publish(h.destQueue, mappers.DeliveryPublishing(msg))
}
