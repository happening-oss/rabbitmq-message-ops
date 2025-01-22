package managers

import (
	"context"
	"log/slog"
	"time"

	"github.com/rabbitmq/amqp091-go"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging"
	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/handlers"
	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/mappers"
	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/selectors"
)

const (
	partialQueueManagementCtxCancelHelpMsg = `Source queue has potentially been partially managed. 
Please check if some messages have been moved from the source queue to temporary queue.
Try to manage queue again and specify the --tempQueue parameter with the currently used temporary queue.
That will cause QueueManager to continue processing from the last processed message that caused error and move all tempQueue messages (also those that were moved to tempQueue during the failed command) to source queue when finished, preserving the order.`
	partialQueueManagementHelpMsg = `Source queue has potentially been partially managed. 
Please check if some messages have been moved from the source queue to temporary queue.
Please check if the last processed message (the one that caused the error, you can do that using "view --count=1") is requeued back to the front of the source queue.
If message is not at the front, please move message to the front of the source queue manually.
If some messages have been moved and last processed message is at the front, try to manage queue again and specify the --tempQueue parameter with the currently used temporary queue.
That will cause QueueManager to continue processing from the last processed message that caused error and move all tempQueue messages (also those that were moved to tempQueue during the failed command) to source queue when finished, preserving the order.
If publishing to the destination queue (move/copy commands) succeeded, but acknowledging the message failed, please manually remove the duplicated message from the source or destination queue.`
	partialTempQueueMoveHelpMsg = `Please manually move remaining messages from the temporary queue to source queue (use "move" command).
Before doing that, please check if the last processed message (the one that caused the error, you can do that using "view --count=1") is requeued back to the front of the temporary queue.
If message is not at the front, please move message to the front of the temporary queue manually.
If publishing to the source queue succeeded, but acknowledging the message failed, please manually remove the duplicated message from the temporary or source queue.`
	partialTempQueueMoveCtxCancelHelpMsg = `Please manually move remaining messages from the temporary queue to source queue (use "move" command).`
)

type QueueManager struct {
	consumer  messaging.Consumer
	log       *slog.Logger
	handler   handlers.MessageHandler
	publisher messaging.Publisher
	selector  selectors.Selector
	tempQueue string
}

func NewQueueManager(consumer messaging.Consumer, log *slog.Logger, handler handlers.MessageHandler, publisher messaging.Publisher, selector selectors.Selector, tempQueue string) *QueueManager {
	return &QueueManager{consumer: consumer, log: log, handler: handler, publisher: publisher, selector: selector, tempQueue: tempQueue}
}

// region Public

func (m *QueueManager) Manage(ctx context.Context, srcQueue string) error {
	messages, err := m.consumer.Consume(srcQueue)
	if err != nil {
		return err
	}

	m.log.Info("processing source queue")

	startTime := time.Now()
	var processedMessages, selectedMessages int
	var lastProcessedMessage amqp091.Delivery

	defer func() {
		m.log.Info("processing source queue finished",
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
				return m.handleMsgProcessingError("error occurred while checking if message is selected", err, msg, srcQueue)
			}

			requeue := true
			if selected {
				selectedMessages++
				// process message with the provided handler
				requeue, err = m.handler.Handle(msg)
				if err != nil {
					return m.handleMsgProcessingError("error occurred while handling message", err, msg, srcQueue)
				}
			}

			if requeue {
				// move/publish message to the temporary queue
				err = m.publisher.Publish(m.tempQueue, mappers.DeliveryPublishing(msg))
				if err != nil {
					return m.handleMsgProcessingError("error occurred while publishing message to temporary queue", err, msg, srcQueue)
				}
			}
			err = msg.Ack(false) // purge/remove message from the source queue
			if err != nil {
				m.logMsgProcessingError("error occurred while acknowledging message", err, msg, srcQueue)
				return err
			}
			if processedMessages%1000 == 0 {
				m.log.Info("processing source queue progress",
					slog.Int("processedMessages", processedMessages),
					slog.Int("selectedMessages", selectedMessages),
					slog.Duration("duration", time.Since(startTime)),
				)
			}
			lastProcessedMessage = msg
		case <-ctx.Done():
			m.log.Error("context cancelled while processing source queue",
				slog.Any("error", ctx.Err()),
				slog.Any("lastProcessedMessage", lastProcessedMessage),
				slog.String("srcQueue", srcQueue),
				slog.String("tempQueue", m.tempQueue),
				slog.String("help", partialQueueManagementCtxCancelHelpMsg),
			)
			return ctx.Err()
		case <-time.After(time.Second):
			break loop
		}
	}

	// move messages back to source queue from temporary queue
	return m.moveTempToSource(ctx, m.tempQueue, srcQueue)
}

func (m *QueueManager) moveTempToSource(ctx context.Context, tempQueue, srcQueue string) error {
	messages, err := m.consumer.Consume(tempQueue)
	if err != nil {
		return err
	}

	m.log.Info("moving messages from temporary to source queue")

	startTime := time.Now()
	var movedMessages int
	var lastMovedMessage amqp091.Delivery

	defer func() {
		m.log.Info("moving messages from temporary to source queue finished",
			slog.Int("movedMessages", movedMessages),
			slog.Duration("duration", time.Since(startTime)),
		)
	}()

loop:
	for {
		select {
		case msg := <-messages:
			// move/publish message back to the source queue
			err = m.publisher.Publish(srcQueue, mappers.DeliveryPublishing(msg))
			if err != nil {
				m.logMoveTempToSrcErr("error occurred while moving message from temporary to source queue", err, msg, srcQueue, tempQueue)
				errReject := msg.Reject(false)
				if errReject != nil {
					m.logMoveTempToSrcErr("error occurred while rejecting message", errReject, msg, srcQueue, tempQueue)
				}
				return err
			}
			movedMessages++
			err = msg.Ack(false)
			if err != nil {
				m.logMoveTempToSrcErr("error occurred while acknowledging message", err, msg, srcQueue, tempQueue)
				return err
			}
			if movedMessages%1000 == 0 {
				m.log.Info("moving messages from temporary to source queue progress",
					slog.Int("movedMessages", movedMessages),
					slog.Duration("duration", time.Since(startTime)),
				)
			}
			lastMovedMessage = msg
		case <-ctx.Done():
			m.log.Error("context cancelled while moving messages from temporary to source queue",
				slog.Any("error", ctx.Err()),
				slog.Any("lastMovedMessage", lastMovedMessage),
				slog.String("srcQueue", srcQueue),
				slog.String("tempQueue", tempQueue),
				slog.String("help", partialTempQueueMoveCtxCancelHelpMsg),
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

func (m *QueueManager) handleMsgProcessingError(errMsg string, err error, msg amqp091.Delivery, srcQueue string) error {
	m.logMsgProcessingError(errMsg, err, msg, srcQueue)
	errReject := msg.Reject(true)
	if errReject != nil {
		m.logMsgProcessingError("error occurred while rejecting message", errReject, msg, srcQueue)
	}
	return err
}

func (m *QueueManager) logMsgProcessingError(errMsg string, err error, msg amqp091.Delivery, srcQueue string) {
	m.log.Error(errMsg,
		slog.Any("error", err),
		slog.Any("msg", msg),
		slog.String("srcQueue", srcQueue),
		slog.String("tempQueue", m.tempQueue),
		slog.String("help", partialQueueManagementHelpMsg),
	)
}

func (m *QueueManager) logMoveTempToSrcErr(errMsg string, err error, msg amqp091.Delivery, srcQueue, tempQueue string) {
	m.log.Error(errMsg,
		slog.Any("error", err),
		slog.Any("msg", msg),
		slog.String("srcQueue", srcQueue),
		slog.String("tempQueue", tempQueue),
		slog.String("help", partialTempQueueMoveHelpMsg),
	)
}

// endregion
