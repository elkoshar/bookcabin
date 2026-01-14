package aggregator

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/elkoshar/bookcabin/api"
	"github.com/elkoshar/bookcabin/pkg/helpers"
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

			flights, err := s.searchProcess(ctxWithTimeout, prov, criteria)

			if err != nil {
				slog.Error(fmt.Sprintf("Provider %s given up: %v", prov.Name(), err))
				errorChan <- err
				return
			}
			resultChan <- flights
		}(p)
	}

	// Wait & Close
	go func() {
		wg.Wait()
		close(resultChan)
		close(errorChan)
	}()

	var allFlights []service.UnifiedFlight
	for res := range resultChan {
		allFlights = append(allFlights, res...)
		providersSucceeded++
	}
	for range errorChan {
		providersFailed++
	}

	for i := range allFlights {
		allFlights[i].Price.Formatted = helpers.FormatIDR(allFlights[i].Price.Amount)

		allFlights[i].Score = calculateScore(allFlights[i])
	}

	sort.Slice(allFlights, func(i, j int) bool {
		return allFlights[i].Score < allFlights[j].Score
	})

	//todo create return flight handling
	if criteria.ReturnDate != "" {
		slog.Info("Round-trip requested")
	}

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

func (s *FlightAggregator) searchProcess(ctx context.Context, p api.FlightProvider, c service.SearchCriteria) ([]service.UnifiedFlight, error) {
	maxRetries := 3
	baseDelay := 100 * time.Millisecond

	var err error
	var flights []service.UnifiedFlight

	for i := 0; i < maxRetries; i++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		flights, err = p.Search(ctx, c)
		if err == nil {
			return flights, nil // Success
		}

		if i < maxRetries-1 {
			delay := baseDelay * time.Duration(math.Pow(2, float64(i)))

			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
				continue
			}
		}
	}
	return nil, err
}

// Formula: (Price / 100.000) * 0.7 + (Duration Hours) * 0.3
func calculateScore(f service.UnifiedFlight) float64 {
	priceFactor := f.Price.Amount / 100000.0
	durationHours := float64(f.Duration.TotalMinutes) / 60.0

	// penalty for every stop / transit
	stopPenalty := float64(f.Stops) * 0.5

	score := (priceFactor * 0.7) + (durationHours * 0.3) + stopPenalty
	return score
}
