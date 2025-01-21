package main

import (
	"github.com/urfave/cli/v2"

	"rabbitmq-message-ops/cmd/cli/util"
	"rabbitmq-message-ops/internal/messaging/management/handlers"
)

func copyMessages() *cli.Command {
	return &cli.Command{
		Name:  "copy",
		Usage: "Copy messages from source to destination queue",
		UsageText: `rabbitmq-cli copy [command options]
Example: rabbitmq-cli -q <srcQueueName> -f 'type == "<some.msg.type>"' copy -d <destQueueName>`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "destination",
				Aliases:  []string{"d"},
				Usage:    "Name of the destination queue to copy messages to.",
				Required: true,
			},
		},
		Action: func(c *cli.Context) error {
			destQueue := c.String("destination")
			// check if destination queue exists
			_, err := util.GetClient(c).GetQueueInfo(destQueue)
			if err != nil {
				return err
			}
			return manageQueue(c, handlers.NewCopyHandler(util.GetPublisher(c), destQueue))
		},
	}
}
