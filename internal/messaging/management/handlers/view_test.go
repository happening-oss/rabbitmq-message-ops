package handlers_test

import (
	"os"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/management/handlers"
)

var _ = Describe("View handler", func() {

	When("writing to stdout", func() {
		It("succeeds", func() {
			handler := handlers.NewViewHandler(1, nil)
			requeue, err := handler.Handle(amqp091.Delivery{
				Headers: map[string]interface{}{"type": "msg.type1"},
				Body:    []byte("body1"),
				Type:    "someType",
			})
			Expect(err).ToNot(HaveOccurred())
			Expect(requeue).To(BeTrue())
		})
	})

	When("writing to output file", func() {

		It("saves messages in the expected format", func() {
			currentTime := time.Unix(1, 0).UTC()
			messages := []amqp091.Delivery{
				{Headers: map[string]interface{}{"type": "msg.type1"}, Body: []byte("body1"), Type: "someType"},
				{Headers: map[string]interface{}{"type": "msg.type2"}, ConsumerTag: "consumerTag"},
				{Headers: map[string]interface{}{"type": "msg.type3", "userID": "user123"}, AppId: "123"},
				{Headers: map[string]interface{}{"type": "msg.type4"}, Body: []byte("body4"), Timestamp: currentTime},
				{Body: []byte("body5")},
				{Headers: map[string]interface{}{"type": "msg.type6"}},
				{Headers: map[string]interface{}{"type": "msg.type7"}},
				{Headers: map[string]interface{}{"type": "msg.type8"}},
			}

			outputFile, err := os.CreateTemp("", "")
			Expect(err).ToNot(HaveOccurred())
			defer func() {
				err = os.Remove(outputFile.Name())
				Expect(err).ToNot(HaveOccurred())
			}()

			handler := handlers.NewViewHandler(5, outputFile)

			for _, msg := range messages {
				requeue, err := handler.Handle(msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(requeue).To(BeTrue())
			}

			data, err := os.ReadFile(outputFile.Name())
			Expect(err).ToNot(HaveOccurred())
			Expect(string(data)).To(Equal(`{"headers":{"type":"msg.type1"},"type":"someType","body":"body1"}
{"headers":{"type":"msg.type2"}}
{"headers":{"type":"msg.type3","userID":"user123"},"appID":"123"}
{"headers":{"type":"msg.type4"},"timestamp":"1970-01-01T00:00:01Z","body":"body4"}
{"body":"body5"}
`))
		})
	})
})
