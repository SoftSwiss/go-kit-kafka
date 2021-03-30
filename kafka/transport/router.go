package transport

import (
	"context"
	"fmt"

	"github.com/alebabai/go-kit-kafka/kafka"
)

// HandlersMapping represents Topic -> []Handler mapping
type HandlersMapping map[string][]kafka.Handler

type Router struct {
	handlersMapping HandlersMapping
}

func NewRouter(opts ...RouterOption) (*Router, error) {
	r := &Router{
		handlersMapping: make(HandlersMapping),
	}

	for _, opt := range opts {
		opt(r)
	}

	return r, nil
}

func (r *Router) AddHandler(topic string, handler kafka.Handler) *Router {
	if len(r.handlersMapping) == 0 {
		r.handlersMapping = make(HandlersMapping)
	}

	r.handlersMapping[topic] = append(r.handlersMapping[topic], handler)

	return r
}

func (r *Router) Handle(ctx context.Context, msg kafka.Message) error {
	for _, h := range r.handlersMapping[msg.Topic()] {
		if err := h.Handle(ctx, msg); err != nil {
			return fmt.Errorf("failed to handle message from kafka topic=%s: %w", msg.Topic(), err)
		}
	}

	return nil
}
