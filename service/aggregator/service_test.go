package aggregator_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/elkoshar/bookcabin/service"
	"github.com/elkoshar/bookcabin/service/aggregator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockProvider implements FlightProvider interface for testing
type MockProvider struct {
	mock.Mock
}

func (m *MockProvider) Name() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockProvider) Search(ctx context.Context, criteria service.SearchCriteria) ([]service.UnifiedFlight, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).([]service.UnifiedFlight), args.Error(1)
}

func TestNewAggregator(t *testing.T) {
	provider1 := &MockProvider{}
	provider2 := &MockProvider{}
	timeout := 5 * time.Second

	agg := aggregator.NewAggregator(timeout, provider1, provider2)

	assert.NotNil(t, agg)
	// Note: cannot test private fields directly from external package
}

func TestFlightAggregator_SearchAll_SameOriginDestination(t *testing.T) {
	provider := &MockProvider{}
	agg := aggregator.NewAggregator(5*time.Second, provider)

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "CGK",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	_, err := agg.SearchAll(ctx, criteria)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "origin and destination cannot be the same")
}

func TestFlightAggregator_SearchAll_DepartOnly_Success(t *testing.T) {
	provider1 := &MockProvider{}
	provider2 := &MockProvider{}

	// Mock provider 1 response
	provider1.On("Search", mock.Anything, mock.Anything).Return([]service.UnifiedFlight{
		{
			ID:           "TEST1",
			Provider:     "Test Provider 1",
			FlightNumber: "TP100",
			Price:        service.PriceInfo{Amount: 500000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 120},
			Stops:        0,
		},
	}, nil)

	// Mock provider 2 response
	provider2.On("Search", mock.Anything, mock.Anything).Return([]service.UnifiedFlight{
		{
			ID:           "TEST2",
			Provider:     "Test Provider 2",
			FlightNumber: "TP200",
			Price:        service.PriceInfo{Amount: 750000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 90},
			Stops:        1,
		},
	}, nil)

	aggregator := aggregator.NewAggregator(5*time.Second, provider1, provider2)

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	result, err := aggregator.SearchAll(ctx, criteria)

	assert.NoError(t, err)
	assert.Equal(t, criteria, result.Criteria)
	assert.Len(t, result.Flights, 2)
	assert.Empty(t, result.ReturnFlights)
	assert.Equal(t, 2, result.Metadata.TotalResults)
	assert.Equal(t, 2, result.Metadata.ProvidersQueried)
	assert.Equal(t, 2, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 0, result.Metadata.ProvidersFailed)

	// Check if flights are sorted by score (lower score first)
	assert.True(t, result.Flights[0].Score <= result.Flights[1].Score)

	provider1.AssertExpectations(t)
	provider2.AssertExpectations(t)
}

func TestFlightAggregator_SearchAll_RoundTrip_Success(t *testing.T) {
	provider := &MockProvider{}

	// Mock provider response for depart flights
	provider.On("Search", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "CGK" && criteria.Destination == "DPS"
	})).Return([]service.UnifiedFlight{
		{
			ID:           "DEPART1",
			Provider:     "Test Provider",
			FlightNumber: "TP100",
			Price:        service.PriceInfo{Amount: 500000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 120},
		},
	}, nil)

	// Mock provider response for return flights
	provider.On("Search", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "DPS" && criteria.Destination == "CGK"
	})).Return([]service.UnifiedFlight{
		{
			ID:           "RETURN1",
			Provider:     "Test Provider",
			FlightNumber: "TP200",
			Price:        service.PriceInfo{Amount: 600000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 130},
		},
	}, nil)

	agg := aggregator.NewAggregator(5*time.Second, provider)

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		ReturnDate:    "2025-12-20",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	assert.NoError(t, err)
	assert.Len(t, result.Flights, 1)
	assert.Len(t, result.ReturnFlights, 1)
	assert.Equal(t, 2, result.Metadata.TotalResults)
	assert.Equal(t, 2, result.Metadata.ProvidersQueried) // 1 provider * 2 searches

	provider.AssertExpectations(t)
}

func TestFlightAggregator_SearchAll_ProviderFailure(t *testing.T) {
	provider1 := &MockProvider{}
	provider2 := &MockProvider{}

	// Provider 1 succeeds
	provider1.On("Name").Return("Test Provider 1").Maybe()
	provider1.On("Search", mock.Anything, mock.Anything).Return([]service.UnifiedFlight{
		{
			ID:           "SUCCESS1",
			Provider:     "Test Provider 1",
			FlightNumber: "TP100",
			Price:        service.PriceInfo{Amount: 500000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 120},
		},
	}, nil)

	// Provider 2 fails
	provider2.On("Name").Return("Test Provider 2").Maybe()
	provider2.On("Search", mock.Anything, mock.Anything).Return([]service.UnifiedFlight(nil), errors.New("provider error"))

	agg := aggregator.NewAggregator(5*time.Second, provider1, provider2)

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	assert.NoError(t, err) // Aggregator should still succeed with partial results
	assert.Len(t, result.Flights, 1)
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 1, result.Metadata.ProvidersFailed)

	// Don't assert expectations since Name() calls might vary
}

func TestFlightAggregator_SearchAll_ContextTimeout(t *testing.T) {
	provider := &MockProvider{}

	// Mock provider to simulate delay that exceeds context timeout
	provider.On("Name").Return("Slow Provider")
	provider.On("Search", mock.Anything, mock.Anything).Return([]service.UnifiedFlight(nil), context.DeadlineExceeded)

	agg := aggregator.NewAggregator(100*time.Millisecond, provider) // Very short timeout

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	assert.NoError(t, err) // Should still return results (even if empty)
	assert.Empty(t, result.Flights)
	assert.Equal(t, 0, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 1, result.Metadata.ProvidersFailed)

	provider.AssertExpectations(t)
}
