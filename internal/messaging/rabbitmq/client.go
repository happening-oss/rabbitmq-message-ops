package rabbitmq

import (
	"fmt"

	rabbithole "github.com/michaelklishin/rabbit-hole/v2"
)

type Client struct {
	client *rabbithole.Client
}

func NewClient(endpoint, username, password string) (*Client, error) {
	rmqc, err := rabbithole.NewClient(endpoint, username, password)
	if err != nil {
		return nil, err
	}
	return &Client{client: rmqc}, nil
}

// region Public

func (c *Client) GetQueueInfo(queue string) (rabbithole.QueueInfo, error) {
	queuesInfo, err := c.client.ListQueues()
	if err != nil {
		return rabbithole.QueueInfo{}, fmt.Errorf("rabbitmq_client: %w", err)
	}
	for _, qInfo := range queuesInfo {
		if qInfo.Name == queue {
			return qInfo, nil
		}
	}
	return rabbithole.QueueInfo{}, fmt.Errorf("rabbitmq_client: queue %v not found", queue)
}

// endregion
