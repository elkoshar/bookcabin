package aggregator

import (
	"time"

	"github.com/elkoshar/bookcabin/api"
)

type FlightAggregator struct {
	providers []api.FlightProvider
	timeout   time.Duration
}

func NewAggregator(timeout int, providers ...api.FlightProvider) *FlightAggregator {
	return &FlightAggregator{
		providers: providers,
		timeout:   time.Duration(timeout) * time.Second,
	}
}
