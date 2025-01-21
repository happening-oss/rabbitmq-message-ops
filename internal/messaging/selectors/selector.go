package selectors

import "github.com/rabbitmq/amqp091-go"

//go:generate mockery --case underscore --name "Selector" --output ./mocks

type Selector interface {
	IsSelected(msg amqp091.Delivery) (isSelected bool, err error)
}
