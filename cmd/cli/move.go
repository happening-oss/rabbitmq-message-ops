package main

import (
	"github.com/urfave/cli/v2"

	"rabbitmq-message-ops/cmd/cli/util"
	"rabbitmq-message-ops/internal/messaging/management/handlers"
)

func moveMessages() *cli.Command {
	return &cli.Command{
		Name:  "move",
		Usage: "Move messages from source to destination queue",
		UsageText: `rabbitmq-cli move [command options]
Example: rabbitmq-cli -q <srcQueueName> -f 'type == "<some.msg.type>"' move -d <destQueueName>`,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:     "destination",
				Aliases:  []string{"d"},
				Usage:    "Name of the destination queue to move messages to.",
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
			return manageQueue(c, handlers.NewMoveHandler(util.GetPublisher(c), destQueue))
		},
	}
}
