package main

import (
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"slices"
	"strings"

	"github.com/rabbitmq/amqp091-go"
	"github.com/urfave/cli/v2"

	"rabbitmq-message-ops/cmd/cli/util"
	"rabbitmq-message-ops/internal/messaging"
	"rabbitmq-message-ops/internal/messaging/management/handlers"
	"rabbitmq-message-ops/internal/messaging/management/managers"
	"rabbitmq-message-ops/internal/messaging/rabbitmq"
	"rabbitmq-message-ops/internal/messaging/selectors"
)

// region Helpers

func manageQueue(c *cli.Context, handler handlers.MessageHandler) error {
	endpoint := c.String("endpoint")
	tempQueue := c.String("temp-queue")
	filterExpr := c.String("filter")
	srcQueue := c.String("queue")

	queueInfo, err := util.GetClient(c).GetQueueInfo(srcQueue)
	if err != nil {
		return err
	}

	// perform additional logic before creating and running manager
	switch queueInfo.Type {
	case amqp091.QueueTypeClassic, amqp091.QueueTypeQuorum:
		// create a temporary queue to preserve the original order of messages in the source queue
		var cleanup func()
		tempQueue, cleanup, err = handleTempQueue(endpoint, tempQueue)
		if err != nil {
			return err
		}
		defer cleanup()
	case amqp091.QueueTypeStream:
		supportedCommands := []string{"view", "copy"}
		if !slices.Contains(supportedCommands, c.Command.Name) {
			return fmt.Errorf("%v queue type does not support %v command. Supported commands: %v", amqp091.QueueTypeStream, c.Command.Name, strings.Join(supportedCommands, ","))
		}
	}

	consumer, err := consumerFactory(queueInfo.Type, endpoint)
	if err != nil {
		return err
	}
	defer func() {
		closeErr := consumer.Close()
		if closeErr != nil {
			log.Error("error while closing consumer", slog.Any("error", closeErr))
		}
	}()

	var selector selectors.Selector
	if filterExpr != "" {
		selector, err = selectors.NewFilterExprSelector(filterExpr)
		if err != nil {
			return err
		}
	} else {
		selector = selectors.NewYesSelector()
	}

	manager, err := managerFactory(queueInfo.Type, consumer, util.GetPublisher(c), handler, selector, tempQueue)
	if err != nil {
		return err
	}

	log.Info("source queue messages info",
		slog.Int("total", queueInfo.Messages),
		slog.Int("ready", queueInfo.MessagesReady),
		slog.Int("unacknowledged", queueInfo.MessagesUnacknowledged),
	)

	return manager.Manage(c.Context, srcQueue)
}

func handleTempQueue(endpoint, queue string) (tempQueue string, cleanup func(), err error) {
	connection, err := amqp091.Dial(endpoint)
	if err != nil {
		return "", nil, err
	}

	closeConnection := func() {
		closeErr := connection.Close()
		if closeErr != nil {
			log.Error("error while closing connection", slog.Any("error", closeErr))
		}
	}
	defer func() {
		if err != nil {
			closeConnection()
		}
	}()

	channel, err := connection.Channel()
	if err != nil {
		return "", nil, err
	}

	if queue == "" {
		tQueue, err := channel.QueueDeclare(queue, true, false, false, false, nil)
		if err != nil {
			return "", nil, err
		}
		queue = tQueue.Name
	} else {
		_, err = channel.QueueDeclarePassive(queue, true, false, false, false, nil)
		if err != nil {
			return "", nil, err
		}
	}

	deleteQueue := func() {
		_, delErr := channel.QueueDelete(queue, false, true, false)
		if delErr != nil {
			log.Warn("failed to delete temporary queue. If needed, please manually delete temporary queue", slog.Any("error", delErr), slog.String("queue", queue))
		}
	}

	return queue, func() {
		deleteQueue()
		closeConnection()
	}, nil
}

// endregion

// region Factory

func consumerFactory(queueType, endpoint string) (messaging.Consumer, error) {
	switch queueType {
	case amqp091.QueueTypeClassic, amqp091.QueueTypeQuorum:
		return rabbitmq.NewSimpleConsumer(endpoint)
	case amqp091.QueueTypeStream:
		return rabbitmq.NewSimpleConsumer(endpoint, "first")
	default:
		return nil, errors.New("unsupported queue type: " + queueType)
	}
}

func managerFactory(queueType string, consumer messaging.Consumer, publisher messaging.Publisher, handler handlers.MessageHandler, selector selectors.Selector, tempQueue string) (managers.Manager, error) {
	switch queueType {
	case amqp091.QueueTypeClassic, amqp091.QueueTypeQuorum:
		return managers.NewQueueManager(consumer, log, handler, publisher, selector, tempQueue), nil
	case amqp091.QueueTypeStream:
		return managers.NewStreamManager(consumer, log, handler, publisher, selector), nil
	default:
		return nil, errors.New("unsupported queue type: " + queueType)
	}
}

// endregion

// region Builders

func buildRabbitMQClient(endpoint, httpAPIEndpoint string) (*rabbitmq.Client, error) {
	url, err := url.Parse(endpoint)
	if err != nil {
		return nil, err
	}
	password, _ := url.User.Password()
	if httpAPIEndpoint == "" {
		httpAPIEndpoint = fmt.Sprintf("http://%v:1%v", url.Hostname(), url.Port())
	}
	return rabbitmq.NewClient(httpAPIEndpoint, url.User.Username(), password)
}

func buildPublisher(endpoint string) (messaging.Publisher, error) {
	return rabbitmq.NewSimplePublisher(endpoint)
}

// endregion
