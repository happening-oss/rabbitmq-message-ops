package main

import (
	"errors"
	"log/slog"
	"os"

	"github.com/urfave/cli/v2"

	"rabbitmq-message-ops/cmd/cli/util"
)

var log *slog.Logger
var levelVar *slog.LevelVar

func init() {
	levelVar = new(slog.LevelVar)
	log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: levelVar}))
}

// region Flags

var flagEndpoint = &cli.StringFlag{
	Name:     "endpoint",
	Usage:    "RabbitMQ server address to connect to.",
	EnvVars:  []string{"RABBITMQ_ENDPOINT"},
	Required: true,
}

var flagHTTPAPIEndpoint = &cli.StringFlag{
	Name:    "http-api-endpoint",
	Usage:   "RabbitMQ HTTP API server address to connect to.",
	EnvVars: []string{"RABBITMQ_HTTP_API_ENDPOINT"},
}

var flagQueue = &cli.StringFlag{
	Name:     "queue",
	Aliases:  []string{"q"},
	Usage:    "Name of the source queue to manage.",
	Required: true,
}

var flagTempQueue = &cli.StringFlag{
	Name:    "temp-queue",
	Aliases: []string{"t"},
	Usage: `Temporary queue is used to preserve the original order of messages when performing commands. All messages that should be kept in the source queue are temporarily moved to temporary queue. After the operation is completed, the messages are moved back to the source queue in the same order.
If provided, queue must already exist. If not provided, a new random queue will be created.`,
}

var flagFilter = &cli.StringFlag{
	Name:    "filter",
	Aliases: []string{"f"},
	Usage:   "Filter messages based on filter expression (https://expr-lang.org/).",
}

var flagVerbosity = &cli.StringFlag{
	Name:    "verbosity",
	Aliases: []string{"v"},
	Usage:   "Logging verbosity level (info, error). Default is error.",
}

func buildCLIApp() *cli.App {
	return &cli.App{
		Name:  "rabbitmq-cli",
		Usage: "Manage rabbitmq queues",
		Flags: []cli.Flag{
			flagEndpoint,
			flagHTTPAPIEndpoint,
			flagQueue,
			flagTempQueue,
			flagFilter,
			flagVerbosity,
		},
		Before: func(ctx *cli.Context) error {
			endpoint := ctx.String("endpoint")
			httpAPIEndpoint := ctx.String("http-api-endpoint")
			verbosity := ctx.String("verbosity")

			levelVar.Set(slog.LevelError) // default verbosity level is "error"
			if verbosity != "" {
				level, err := verbosityLevel(verbosity)
				if err != nil {
					return err
				}
				levelVar.Set(level)
			}

			client, err := buildRabbitMQClient(endpoint, httpAPIEndpoint)
			if err != nil {
				return err
			}
			util.AttachClient(ctx, client)

			publisher, err := buildPublisher(endpoint)
			if err != nil {
				return err
			}
			util.AttachPublisher(ctx, publisher)

			return nil
		},
		After: func(ctx *cli.Context) error {
			publisher := util.GetPublisher(ctx)
			if publisher == nil {
				return nil
			}
			closeErr := publisher.Close()
			if closeErr != nil {
				log.Error("error while closing publisher", slog.Any("error", closeErr))
			}
			return nil
		},
		Commands: []*cli.Command{
			viewMessages(),
			moveMessages(),
			copyMessages(),
			purgeMessages(),
		},
	}
}

func verbosityLevel(verbosity string) (slog.Level, error) {
	switch verbosity {
	case "info":
		return slog.LevelInfo, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, errors.New("unsupported verbosity level")
	}
}
