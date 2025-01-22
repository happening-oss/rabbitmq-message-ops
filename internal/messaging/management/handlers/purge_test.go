package handlers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/handlers"
)

var _ = Describe("Purge handler", func() {
	var handler *handlers.PurgeHandler

	BeforeEach(func() {
		handler = handlers.NewPurgeHandler()
	})

	Describe("handling message", func() {
		It("doesn't requeue message", func() {
			requeue, err := handler.Handle(amqp091.Delivery{})
			Expect(err).ToNot(HaveOccurred())
			Expect(requeue).To(BeFalse())
		})
	})
})
