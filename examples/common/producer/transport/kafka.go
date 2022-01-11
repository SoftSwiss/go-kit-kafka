package transport

import (
	"context"
	"github.com/SoftSwiss/go-kit-kafka/kafka/tracing"

	"github.com/SoftSwiss/go-kit-kafka/kafka"
	"github.com/SoftSwiss/go-kit-kafka/kafka/transport"

	"github.com/SoftSwiss/go-kit-kafka/examples/common/producer/endpoint"
)

func NewKafkaProducer(handler kafka.Handler, topic string) *transport.Producer {
	return transport.NewProducer(
		handler,
		topic,
		encodeProduceEventRequest,
		transport.ProducerBefore(tracing.MessageToContext),
	)
}

func encodeProduceEventRequest(ctx context.Context, msg *kafka.Message, request interface{}) error {
	req := request.(endpoint.ProduceEventRequest)
	return transport.EncodeJSONRequest(ctx, msg, req.Payload)
}
