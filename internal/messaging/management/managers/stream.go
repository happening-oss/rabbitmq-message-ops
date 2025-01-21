package managers

import (
	"context"
	"log/slog"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"rabbitmq-message-ops/internal/messaging"
	"rabbitmq-message-ops/internal/messaging/management/handlers"
	"rabbitmq-message-ops/internal/messaging/selectors"
)

const (
	partialStreamManagementHelpMsg = "Source stream has potentially been partially managed. Please act accordingly on the destination queue."
)

type StreamManager struct {
	consumer  messaging.Consumer
	log       *slog.Logger
	handler   handlers.MessageHandler
	publisher messaging.Publisher
	selector  selectors.Selector
}

func NewStreamManager(consumer messaging.Consumer, log *slog.Logger, handler handlers.MessageHandler, publisher messaging.Publisher, selector selectors.Selector) *StreamManager {
	return &StreamManager{consumer: consumer, log: log, handler: handler, publisher: publisher, selector: selector}
}

// region Public

func (m *StreamManager) Manage(ctx context.Context, srcStream string) error {
	messages, err := m.consumer.Consume(srcStream)
	if err != nil {
		return err
	}

	m.log.Info("processing source stream")

	startTime := time.Now()
	var processedMessages, selectedMessages int
	var lastProcessedMessage amqp091.Delivery

	defer func() {
		m.log.Info("processing source stream finished",
			slog.Int("processedMessages", processedMessages),
			slog.Int("selectedMessages", selectedMessages),
			slog.Duration("duration", time.Since(startTime)),
		)
	}()

loop:
	for {
		select {
		case msg := <-messages:
			processedMessages++
			selected, err := m.selector.IsSelected(msg)
			if err != nil {
				m.logHandleSrcMsgErr("error occurred while checking if message is selected", err, msg, srcStream)
				errReject := msg.Reject(true)
				if errReject != nil {
					m.logHandleSrcMsgErr("error occurred while rejecting message", errReject, msg, srcStream)
				}
				return err
			}
			if selected {
				selectedMessages++
				_, err = m.handler.Handle(msg)
				if err != nil {
					m.logHandleSrcMsgErr("error occurred while handling message", err, msg, srcStream)
					errReject := msg.Reject(true)
					if errReject != nil {
						m.logHandleSrcMsgErr("error occurred while rejecting message", errReject, msg, srcStream)
					}
					return err
				}
			}
			err = msg.Ack(false) // purge/remove message from the source queue
			if err != nil {
				m.logHandleSrcMsgErr("error occurred while acknowledging message", err, msg, srcStream)
				return err
			}
			if processedMessages%1000 == 0 {
				m.log.Info("processing source stream progress",
					slog.Int("processedMessages", processedMessages),
					slog.Int("selectedMessages", selectedMessages),
					slog.Duration("duration", time.Since(startTime)),
				)
			}
			lastProcessedMessage = msg
		case <-ctx.Done():
			m.log.Error("context cancelled while processing source stream",
				slog.Any("error", ctx.Err()),
				slog.Any("lastProcessedMessage", lastProcessedMessage),
				slog.String("srcStream", srcStream),
				slog.String("help", partialStreamManagementHelpMsg),
			)
			return ctx.Err()
		case <-time.After(time.Second):
			break loop
		}
	}

	return nil
}

// endregion

// region Private

func (m *StreamManager) logHandleSrcMsgErr(errMsg string, err error, msg amqp091.Delivery, srcStream string) {
	m.log.Error(errMsg,
		slog.Any("error", err),
		slog.Any("msg", msg),
		slog.String("srcStream", srcStream),
		slog.String("help", partialStreamManagementHelpMsg),
	)
}

// endregion
