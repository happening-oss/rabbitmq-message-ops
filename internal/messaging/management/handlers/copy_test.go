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

var _ = Describe("Copy handler", func() {
	var pubMock *mmocks.Publisher

	BeforeEach(func() {
		pubMock = mmocks.NewPublisher(GinkgoT())
	})

	var handler *handlers.CopyHandler

	BeforeEach(func() {
		handler = handlers.NewCopyHandler(pubMock, "destQueue")
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

			It("requeues message", func() {
				requeue, err := handler.Handle(amqp091.Delivery{})

				Expect(err).ToNot(HaveOccurred())
				Expect(requeue).To(BeTrue())
			})
		})
	})
})
