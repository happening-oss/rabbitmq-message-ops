package selectors

import (
	"time"
	"unicode/utf8"

	"github.com/rabbitmq/amqp091-go"
)

func SubsetFromDelivery(msg amqp091.Delivery) DeliverySubset {
	var timestamp string
	var body any

	if !msg.Timestamp.IsZero() {
		timestamp = msg.Timestamp.Format(time.RFC3339Nano)
	}

	if len(msg.Body) > 0 {
		if utf8.Valid(msg.Body) {
			body = string(msg.Body)
		} else {
			body = msg.Body
		}
	}

	return DeliverySubset{
		Headers:         msg.Headers,
		ContentType:     msg.ContentType,
		ContentEncoding: msg.ContentEncoding,
		DeliveryMode:    msg.DeliveryMode,
		Priority:        msg.Priority,
		CorrelationID:   msg.CorrelationId,
		ReplyTo:         msg.ReplyTo,
		Expiration:      msg.Expiration,
		MessageID:       msg.MessageId,
		Timestamp:       timestamp,
		Type:            msg.Type,
		UserID:          msg.UserId,
		AppID:           msg.AppId,
		Redelivered:     msg.Redelivered,
		Exchange:        msg.Exchange,
		RoutingKey:      msg.RoutingKey,
		Body:            body,
	}
}

// region Structs

type DeliverySubset struct {
	Headers         amqp091.Table `json:"headers,omitempty" expr:"headers"`
	ContentType     string        `json:"contentType,omitempty" expr:"contentType"`
	ContentEncoding string        `json:"contentEncoding,omitempty" expr:"contentEncoding"`
	DeliveryMode    uint8         `json:"deliveryMode,omitempty" expr:"deliveryMode"`
	Priority        uint8         `json:"priority,omitempty" expr:"priority"`
	CorrelationID   string        `json:"correlationID,omitempty" expr:"correlationID"`
	ReplyTo         string        `json:"replyTo,omitempty" expr:"replyTo"`
	Expiration      string        `json:"expiration,omitempty" expr:"expiration"`
	MessageID       string        `json:"messageID,omitempty" expr:"messageID"`
	Timestamp       string        `json:"timestamp,omitempty" expr:"timestamp"`
	Type            string        `json:"type,omitempty" expr:"type"`
	UserID          string        `json:"userID,omitempty" expr:"userID"`
	AppID           string        `json:"appID,omitempty" expr:"appID"`
	Redelivered     bool          `json:"redelivered,omitempty" expr:"redelivered"`
	Exchange        string        `json:"exchange,omitempty" expr:"exchange"`
	RoutingKey      string        `json:"routingKey,omitempty" expr:"routingKey"`
	Body            any           `json:"body,omitempty" expr:"-"`
}

// endregion
