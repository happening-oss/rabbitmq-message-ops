package util

import (
	"context"

	"github.com/urfave/cli/v2"

	"rabbitmq-message-ops/internal/messaging"
	"rabbitmq-message-ops/internal/messaging/rabbitmq"
)

type ctxKey string

const (
	clientKey    ctxKey = "rabbitmq-client"
	publisherKey ctxKey = "rabbitmq-publisher"
)

func GetClient(ctx *cli.Context) *rabbitmq.Client {
	client := ctx.Context.Value(clientKey)
	if client == nil {
		return nil
	}
	return client.(*rabbitmq.Client)
}

func AttachClient(ctx *cli.Context, client *rabbitmq.Client) {
	ctx.Context = context.WithValue(ctx.Context, clientKey, client)
}

func GetPublisher(ctx *cli.Context) messaging.Publisher {
	publisher := ctx.Context.Value(publisherKey)
	if publisher == nil {
		return nil
	}
	return publisher.(messaging.Publisher)
}

func AttachPublisher(ctx *cli.Context, publisher messaging.Publisher) {
	ctx.Context = context.WithValue(ctx.Context, publisherKey, publisher)
}
