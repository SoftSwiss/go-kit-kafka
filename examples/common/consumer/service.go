package consumer

import (
	"context"

	"github.com/SoftSwiss/go-kit-kafka/examples/common/domain"
)

type Service interface {
	Create(ctx context.Context, e domain.Event) error
	List(ctx context.Context) ([]domain.Event, error)
}
