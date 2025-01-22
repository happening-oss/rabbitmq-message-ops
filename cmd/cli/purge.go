package main

import (
	"github.com/urfave/cli/v2"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/handlers"
)

func purgeMessages() *cli.Command {
	return &cli.Command{
		Name:  "purge",
		Usage: "Purge messages from the queue",
		UsageText: `rabbitmq-cli purge [command options]
Example: rabbitmq-cli -q <srcQueueName> -f 'type == "<some.msg.type>"' purge`,
		Action: func(c *cli.Context) error {
			return manageQueue(c, handlers.NewPurgeHandler())
		},
	}
}
