package endpoint

import (
	"github.com/SoftSwiss/go-kit-kafka/examples/common/domain"
)

type CreateEventRequest struct {
	Payload *domain.Event
}

type CreateEventResponse struct {
}

type ListEventsRequest struct {
}

type ListEventsResponse struct {
	Results []domain.Event
}
