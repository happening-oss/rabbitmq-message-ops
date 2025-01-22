package handlers

import (
	"encoding/json"
	"os"

	"github.com/rabbitmq/amqp091-go"

	"github.com/happening-oss/rabbitmq-message-ops/internal/messaging/selectors"
)

type ViewHandler struct {
	count      int
	outputFile *os.File
}

func NewViewHandler(count int, outputFile *os.File) *ViewHandler {
	return &ViewHandler{count: count, outputFile: outputFile}
}

func (h *ViewHandler) Handle(msg amqp091.Delivery) (bool, error) {
	// if we viewed --count messages, return
	if h.count <= 0 {
		return true, nil
	}

	var encoder *json.Encoder
	if h.outputFile != nil {
		encoder = json.NewEncoder(h.outputFile)
	} else {
		encoder = json.NewEncoder(os.Stdout)
	}

	err := encoder.Encode(selectors.SubsetFromDelivery(msg)) // write message to file/stdout for viewing
	if err != nil {
		return true, err
	}

	h.count--

	return true, nil
}
