package aggregator

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"sync"
	"time"

	"github.com/elkoshar/bookcabin/api"
	"github.com/elkoshar/bookcabin/service"
)

type FlightAggregator struct {
	providers []api.FlightProvider
	timeout   time.Duration
}

func NewAggregator(timeout time.Duration, providers ...api.FlightProvider) *FlightAggregator {
	return &FlightAggregator{
		providers: providers,
		timeout:   timeout,
	}
}

func (s *FlightAggregator) Search(ctx context.Context, criteria service.SearchCriteria) (service.SearchResponse, error) {
	startTime := time.Now()

	if criteria.Origin == criteria.Destination {
		return service.SearchResponse{}, fmt.Errorf("origin and destination cannot be the same")
	}

	resultChan := make(chan []service.UnifiedFlight, len(s.providers))
	errorChan := make(chan error, len(s.providers))
	var wg sync.WaitGroup

	ctxWithTimeout, cancel := context.WithTimeout(ctx, s.timeout)
	defer cancel()

	providersSucceeded := 0
	providersFailed := 0

	for _, p := range s.providers {
		wg.Add(1)
		go func(prov api.FlightProvider) {
			defer wg.Done()

			flights, err := prov.Search(ctxWithTimeout, criteria)
			if err != nil {
				fmt.Printf("❌ Provider %s Error: %v\n", prov.Name(), err)
				slog.Error(fmt.Sprintf("Provider %s failed: %v", prov.Name(), err))
				errorChan <- err
				return
			}
			resultChan <- flights
		}(p)
	}

	// WaitGroup Closer
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	// Collect Data
	var allFlights []service.UnifiedFlight
	for res := range resultChan {
		allFlights = append(allFlights, res...)
		providersSucceeded++
	}
	for range errorChan {
		providersFailed++
	}

	// Sorting Logic (Best Value)
	for i := range allFlights {
		priceScore := allFlights[i].Price.Amount / 100000
		durScore := float64(allFlights[i].Duration.TotalMinutes) / 60
		allFlights[i].Score = (priceScore * 0.7) + (durScore * 0.3)
	}
	sort.Slice(allFlights, func(i, j int) bool {
		return allFlights[i].Score < allFlights[j].Score
	})

	return service.SearchResponse{
		Criteria: criteria,
		Metadata: service.Metadata{
			TotalResults:       len(allFlights),
			ProvidersQueried:   len(s.providers),
			ProvidersSucceeded: providersSucceeded,
			ProvidersFailed:    providersFailed,
			SearchTimeMs:       time.Since(startTime).Milliseconds(),
		},
		Flights: allFlights,
	}, nil
}
