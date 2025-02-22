package transport

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/SoftSwiss/go-kit-kafka/kafka"
	"github.com/SoftSwiss/go-kit-kafka/kafka/transport"

	"github.com/SoftSwiss/go-kit-kafka/examples/common/consumer/endpoint"
	"github.com/SoftSwiss/go-kit-kafka/examples/common/domain"
)

func NewKafkaHandler(e endpoint.Endpoints) kafka.Handler {
	return transport.NewConsumer(
		e.CreateEventEndpoint,
		decodeCreateEventRequest,
	)
}

func decodeCreateEventRequest(_ context.Context, msg *kafka.Message) (interface{}, error) {
	var e domain.Event
	if err := json.Unmarshal(msg.Value, &e); err != nil {
		return nil, fmt.Errorf("failed to unmarshal create event request")
	}

	return endpoint.CreateEventRequest{
		Payload: &e,
	}, nil
}
