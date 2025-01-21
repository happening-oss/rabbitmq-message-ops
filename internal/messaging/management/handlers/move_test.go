package handlers_test

import (
	"errors"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"
	"github.com/stretchr/testify/mock"

	mmocks "rabbitmq-message-ops/internal/messaging/mocks"
	"rabbitmq-message-ops/internal/tests/util"

	"rabbitmq-message-ops/internal/messaging/management/handlers"
)

var _ = Describe("Move handler", func() {
	var pubMock *mmocks.Publisher
	var handler *handlers.MoveHandler

	BeforeEach(func() {
		pubMock = mmocks.NewPublisher(GinkgoT())
		handler = handlers.NewMoveHandler(pubMock, "destQueue")
	})

	Describe("handling message", func() {

		When("publisher throws error", func() {
			BeforeEach(func() {
				pubMock.On(util.NameOf(pubMock.Publish), "destQueue", mock.Anything).Return(errors.New("")).Once()
			})

			It("returns error", func() {
				_, err := handler.Handle(amqp091.Delivery{})

				Expect(err).To(HaveOccurred())
			})
		})

		When("publisher publishes successfully", func() {
			BeforeEach(func() {
				pubMock.On(util.NameOf(pubMock.Publish), "destQueue", mock.Anything).Return(nil).Once()
			})

			It("doesn't requeue message", func() {
				requeue, err := handler.Handle(amqp091.Delivery{})

				Expect(err).ToNot(HaveOccurred())
				Expect(requeue).To(BeFalse())
			})
		})
	})
})
