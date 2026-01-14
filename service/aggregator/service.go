package aggregator

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
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

func (s *FlightAggregator) SearchAll(ctx context.Context, criteria service.SearchCriteria) (service.SearchResponse, error) {
	startTime := time.Now()

	if criteria.Origin == criteria.Destination {
		return service.SearchResponse{}, fmt.Errorf("origin and destination cannot be the same")
	}

	slog.Info(fmt.Sprintf("[Aggregator] Searching DEPART: %s -> %s on %s", criteria.Origin, criteria.Destination, criteria.DepartureDate))
	departFlights, departMeta, err := s.executeSearch(ctx, criteria)
	if err != nil {
		return service.SearchResponse{}, err
	}

	var returnFlights []service.UnifiedFlight
	var returnMeta service.Metadata

	if criteria.ReturnDate != "" {
		slog.Info(fmt.Sprintf("[Aggregator] Searching RETURN: %s -> %s on %s", criteria.Destination, criteria.Origin, criteria.ReturnDate))

		returnCriteria := criteria
		returnCriteria.Origin = criteria.Destination
		returnCriteria.Destination = criteria.Origin
		returnCriteria.DepartureDate = criteria.ReturnDate
		returnFlights, returnMeta, err = s.executeSearch(ctx, returnCriteria)
		if err != nil {
			slog.Error(fmt.Sprintf("Failed to fetch return flights: %v", err))
		}
	}

	totalResults := len(departFlights) + len(returnFlights)
	totalQueried := len(s.providers)
	if criteria.ReturnDate != "" {
		totalQueried = len(s.providers) * 2
	}

	finalMetadata := service.Metadata{
		TotalResults:       totalResults,
		ProvidersQueried:   totalQueried,
		ProvidersSucceeded: departMeta.ProvidersSucceeded + returnMeta.ProvidersSucceeded,
		ProvidersFailed:    departMeta.ProvidersFailed + returnMeta.ProvidersFailed,
		SearchTimeMs:       time.Since(startTime).Milliseconds(),
	}

	return service.SearchResponse{
		Criteria:      criteria,
		Metadata:      finalMetadata,
		Flights:       departFlights,
		ReturnFlights: returnFlights,
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

func (s *FlightAggregator) executeSearch(ctx context.Context, criteria service.SearchCriteria) ([]service.UnifiedFlight, service.Metadata, error) {
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

			time.Sleep(time.Duration(rand.Intn(50)) * time.Millisecond)

			flights, err := s.searchProcess(ctxWithTimeout, prov, criteria)
			if err != nil {
				slog.Error(fmt.Sprintf("Provider %s failed: %v", prov.Name(), err))
				errorChan <- err
				return
			}
			resultChan <- flights
		}(p)
	}

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

	meta := service.Metadata{
		ProvidersSucceeded: providersSucceeded,
		ProvidersFailed:    providersFailed,
	}

	return allFlights, meta, nil
}
