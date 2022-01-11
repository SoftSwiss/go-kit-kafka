package endpoint

import (
	"github.com/SoftSwiss/go-kit-kafka/examples/common/domain"
)

type GenerateEventRequest struct {
}

type GenerateEventResponse struct {
	Result *domain.Event
}

type ProduceEventRequest struct {
	Payload *domain.Event
}

type ProduceEventResponse struct {
}
