package api

import (
	"context"

	"github.com/elkoshar/bookcabin/service"
)

type FlightProvider interface {
	Name() string
	Search(ctx context.Context, criteria service.SearchCriteria) ([]service.UnifiedFlight, error)
}

type FlightAggregator interface {
	SearchAll(ctx context.Context, criteria service.SearchCriteria) (service.SearchResponse, error)
}
