package main

import (
	"log/slog"
	"math"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/handlers"
)

func viewMessages() *cli.Command {
	return &cli.Command{
		Name:        "view",
		Usage:       "View messages from the queue",
		Description: "Views messages from the specified queue preserving the original order of the queue.",
		UsageText: `rabbitmq-cli view [command options]
Example: rabbitmq-cli -q <srcQueueName> -f 'type == "<some.msg.type>"' view -c 5 -o output.txt`,
		Flags: []cli.Flag{
			&cli.IntFlag{
				Name:    "count",
				Aliases: []string{"c"},
				Usage: `Number of messages to view.
Please keep in mind that all messages will be processed and --count only determines up to how many messages will be printed to stdout/file.
If --count parameter is set to <= 0, all messages will be viewed. Default value is 0.`,
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Usage:   "Output file to store viewed messages. If not specified, viewed messages will be printed to stdout.",
			},
		},
		Action: func(c *cli.Context) error {
			count := c.Int("count")
			output := c.String("output")

			if count <= 0 {
				count = math.MaxInt
			}

			var outputFile *os.File
			var err error

			if output != "" {
				outputFile, err = os.Create(output)
				if err != nil {
					return err
				}
				defer func() {
					closeErr := outputFile.Close()
					if closeErr != nil {
						log.Error("error while closing output file", slog.Any("error", closeErr))
					}
				}()
			}

			return manageQueue(c, handlers.NewViewHandler(count, outputFile))
		},
	}
}
