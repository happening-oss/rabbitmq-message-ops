package selectors

import "github.com/rabbitmq/amqp091-go"

type YesSelector struct{}

func NewYesSelector() *YesSelector {
	return &YesSelector{}
}

func (s *YesSelector) IsSelected(_ amqp091.Delivery) (bool, error) {
	return true, nil
}
