package managers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
)

func TestHandlers(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Queue managers")
}

// region Helpers

func initReadChannel(srcMessages []amqp091.Delivery) <-chan amqp091.Delivery {
	srcQueue := make(chan amqp091.Delivery, len(srcMessages))
	for _, msg := range srcMessages {
		srcQueue <- msg
	}
	return srcQueue
}

// endregion
