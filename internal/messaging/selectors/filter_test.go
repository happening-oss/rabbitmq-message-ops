package selectors_test

import (
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rabbitmq/amqp091-go"

	"rabbitmq-message-ops/internal/messaging/selectors"
)

func TestFilters(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Filter tests")
}

var _ = Describe("Filtering", func() {

	When("invalid filter expression", func() {

		It("returns error when empty", func() {
			_, err := selectors.NewFilterExprSelector("")
			Expect(err).To(HaveOccurred())
		})

		It("returns error when non-filterable field", func() {
			expressions := []string{`missing.type == ""`, `Type == "some.msg.type"`, `ConsumerTag == ""`, `body == ""`}
			for _, expr := range expressions {
				_, err := selectors.NewFilterExprSelector(expr)
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(ContainSubstring("unknown name"))
			}
		})
	})

	When("invalid return type", func() {

		expressions := []string{"1 + 2", `"something"`}

		It("returns error while evaluating", func() {
			for _, expr := range expressions {
				selector, err := selectors.NewFilterExprSelector(expr)
				Expect(err).ToNot(HaveOccurred())
				_, err = selector.IsSelected(amqp091.Delivery{})
				Expect(err).To(HaveOccurred())
			}
		})
	})

	When("running table tests", func() {

		It("returns expected results", func() {

			tests := []struct {
				filterExpr string
				msg        amqp091.Delivery
				result     bool
			}{
				{
					`type matches "^msg.*"`,
					amqp091.Delivery{
						Type: "msg.type1",
					},
					true,
				},
				{
					`type matches "^msg.*"`,
					amqp091.Delivery{
						Type: "some.msg.type",
					},
					false,
				},
				{
					`headers.someField matches "someValue.*" and correlationID == "123"`,
					amqp091.Delivery{
						Headers: map[string]interface{}{
							"someField": "someValue123",
						},
						CorrelationId: "123",
					},
					true,
				},
				{
					`timestamp == "2024-02-03T15:04:05.999999999Z"`,
					amqp091.Delivery{
						Timestamp: time.Date(2024, 2, 3, 15, 4, 5, 999999999, time.UTC),
					},
					true,
				},
			}

			for _, test := range tests {
				selector, err := selectors.NewFilterExprSelector(test.filterExpr)
				Expect(err).ToNot(HaveOccurred())
				result, err := selector.IsSelected(test.msg)
				Expect(err).ToNot(HaveOccurred())
				Expect(result).To(Equal(test.result))
			}
		})
	})
})
