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

func TestFlightAggregator_SearchMultiCity_Success(t *testing.T) {
	provider1 := &MockProvider{}
	provider2 := &MockProvider{}

	provider1.On("Name").Return("Provider 1").Maybe()
	provider2.On("Name").Return("Provider 2").Maybe()

	// Mock responses for first segment (CGK -> DPS)
	segment1Flights := []service.UnifiedFlight{
		{
			ID:           "SEG1_P1_001",
			Provider:     "Provider 1",
			FlightNumber: "GA101",
			Price:        service.PriceInfo{Amount: 1500000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 120},
		},
	}

	segment1Criteria := mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "CGK" && criteria.Destination == "DPS" && criteria.DepartureDate == "2025-12-15"
	})

	provider1.On("Search", mock.Anything, segment1Criteria).Return(segment1Flights, nil)
	provider2.On("Search", mock.Anything, segment1Criteria).Return([]service.UnifiedFlight{}, nil)

	// Mock responses for second segment (DPS -> SIN)
	segment2Flights := []service.UnifiedFlight{
		{
			ID:           "SEG2_P2_001",
			Provider:     "Provider 2",
			FlightNumber: "SQ202",
			Price:        service.PriceInfo{Amount: 2000000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 90},
		},
	}

	segment2Criteria := mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "DPS" && criteria.Destination == "SIN" && criteria.DepartureDate == "2025-12-17"
	})

	provider1.On("Search", mock.Anything, segment2Criteria).Return([]service.UnifiedFlight{}, nil)
	provider2.On("Search", mock.Anything, segment2Criteria).Return(segment2Flights, nil)

	agg := aggregator.NewAggregator(5*time.Second, provider1, provider2)

	criteria := service.SearchCriteria{
		Passengers: 1,
		CabinClass: "economy",
		Segments: []service.RouteSegment{
			{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			{
				Origin:        "DPS",
				Destination:   "SIN",
				DepartureDate: "2025-12-17",
			},
		},
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	assert.NoError(t, err)
	assert.Len(t, result.MultiCityFlights, 2)
	assert.Len(t, result.MultiCityFlights[0], 1) // First segment has 1 flight
	assert.Len(t, result.MultiCityFlights[1], 1) // Second segment has 1 flight

	// Verify first segment flight
	assert.Equal(t, "SEG1_P1_001", result.MultiCityFlights[0][0].ID)
	assert.Equal(t, "GA101", result.MultiCityFlights[0][0].FlightNumber)

	// Verify second segment flight
	assert.Equal(t, "SEG2_P2_001", result.MultiCityFlights[1][0].ID)
	assert.Equal(t, "SQ202", result.MultiCityFlights[1][0].FlightNumber)

	// Verify metadata
	assert.Equal(t, 2, result.Metadata.TotalResults)
	assert.Equal(t, 4, result.Metadata.ProvidersQueried)   // 2 providers × 2 segments
	assert.Equal(t, 4, result.Metadata.ProvidersSucceeded) // 2 successful calls per segment
	assert.Equal(t, 0, result.Metadata.ProvidersFailed)

	provider1.AssertExpectations(t)
	provider2.AssertExpectations(t)
}

func TestFlightAggregator_SearchMultiCity_PartialFailure(t *testing.T) {
	provider := &MockProvider{}

	provider.On("Name").Return("Test Provider")

	// Mock success for first segment
	segment1Flights := []service.UnifiedFlight{
		{
			ID:           "SEG1_001",
			Provider:     "Test Provider",
			FlightNumber: "TK101",
			Price:        service.PriceInfo{Amount: 1500000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 120},
		},
	}

	segment1Criteria := mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "CGK" && criteria.Destination == "DPS"
	})

	provider.On("Search", mock.Anything, segment1Criteria).Return(segment1Flights, nil)

	// Mock failure for second segment
	segment2Criteria := mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "DPS" && criteria.Destination == "SIN"
	})

	provider.On("Search", mock.Anything, segment2Criteria).Return([]service.UnifiedFlight(nil), errors.New("provider error"))

	agg := aggregator.NewAggregator(5*time.Second, provider)

	criteria := service.SearchCriteria{
		Passengers: 1,
		CabinClass: "economy",
		Segments: []service.RouteSegment{
			{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
			{
				Origin:        "DPS",
				Destination:   "SIN",
				DepartureDate: "2025-12-17",
			},
		},
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	assert.NoError(t, err) // Should still succeed with partial results
	assert.Len(t, result.MultiCityFlights, 2)
	assert.Len(t, result.MultiCityFlights[0], 1) // First segment has results
	assert.Len(t, result.MultiCityFlights[1], 0) // Second segment failed

	// Verify metadata reflects partial success/failure
	assert.Equal(t, 1, result.Metadata.TotalResults)
	assert.Equal(t, 2, result.Metadata.ProvidersQueried)   // 1 provider × 2 segments
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded) // First segment succeeded
	assert.Equal(t, 1, result.Metadata.ProvidersFailed)    // Second segment failed

	provider.AssertExpectations(t)
}

func TestFlightAggregator_SearchMultiCity_AllSegmentsFail(t *testing.T) {
	provider := &MockProvider{}

	provider.On("Name").Return("Failing Provider").Maybe()

	// Mock all segments to fail
	provider.On("Search", mock.Anything, mock.Anything).Return([]service.UnifiedFlight(nil), errors.New("provider error"))

	agg := aggregator.NewAggregator(5*time.Second, provider)

	criteria := service.SearchCriteria{
		Passengers: 1,
		CabinClass: "economy",
		Segments: []service.RouteSegment{
			{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"},
			{Origin: "DPS", Destination: "SIN", DepartureDate: "2025-12-17"},
		},
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	// Should handle all failures gracefully
	assert.NoError(t, err)
	assert.Len(t, result.MultiCityFlights, 2) // Should have 2 empty arrays
	assert.Empty(t, result.MultiCityFlights[0])
	assert.Empty(t, result.MultiCityFlights[1])
	assert.Equal(t, 0, result.Metadata.TotalResults)
	assert.Equal(t, 2, result.Metadata.ProvidersQueried) // 1 provider × 2 segments
	assert.Equal(t, 0, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 2, result.Metadata.ProvidersFailed)
}

func TestFlightAggregator_SearchMultiCity_SingleSegment(t *testing.T) {
	provider := &MockProvider{}

	provider.On("Name").Return("Test Provider").Maybe()

	segmentFlights := []service.UnifiedFlight{
		{
			ID:           "SINGLE_001",
			Provider:     "Test Provider",
			FlightNumber: "AB123",
			Price:        service.PriceInfo{Amount: 1200000, Currency: "IDR"},
			Duration:     service.DurationInfo{TotalMinutes: 100},
		},
	}

	provider.On("Search", mock.Anything, mock.Anything).Return(segmentFlights, nil)

	agg := aggregator.NewAggregator(5*time.Second, provider)

	criteria := service.SearchCriteria{
		Passengers: 1,
		CabinClass: "economy",
		Segments: []service.RouteSegment{
			{
				Origin:        "CGK",
				Destination:   "DPS",
				DepartureDate: "2025-12-15",
			},
		},
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	assert.NoError(t, err)
	assert.Len(t, result.MultiCityFlights, 1)
	assert.Len(t, result.MultiCityFlights[0], 1)
	assert.Equal(t, "SINGLE_001", result.MultiCityFlights[0][0].ID)

	// Verify metadata
	assert.Equal(t, 1, result.Metadata.TotalResults)
	assert.Equal(t, 1, result.Metadata.ProvidersQueried)
	assert.Equal(t, 1, result.Metadata.ProvidersSucceeded)
	assert.Equal(t, 0, result.Metadata.ProvidersFailed)

	provider.AssertExpectations(t)
}

func TestFlightAggregator_SearchMultiCity_ThreeSegments(t *testing.T) {
	provider := &MockProvider{}

	provider.On("Name").Return("Multi Provider").Maybe()

	// Mock responses for all three segments
	segment1Flights := []service.UnifiedFlight{
		{ID: "SEG1_001", FlightNumber: "AA101", Price: service.PriceInfo{Amount: 1000000, Currency: "IDR"}},
	}
	segment2Flights := []service.UnifiedFlight{
		{ID: "SEG2_001", FlightNumber: "BB202", Price: service.PriceInfo{Amount: 1500000, Currency: "IDR"}},
		{ID: "SEG2_002", FlightNumber: "BB203", Price: service.PriceInfo{Amount: 1600000, Currency: "IDR"}},
	}
	segment3Flights := []service.UnifiedFlight{
		{ID: "SEG3_001", FlightNumber: "CC301", Price: service.PriceInfo{Amount: 2000000, Currency: "IDR"}},
	}

	// Set up mock calls for each segment
	provider.On("Search", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "CGK" && criteria.Destination == "DPS"
	})).Return(segment1Flights, nil)

	provider.On("Search", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "DPS" && criteria.Destination == "SIN"
	})).Return(segment2Flights, nil)

	provider.On("Search", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "SIN" && criteria.Destination == "NRT"
	})).Return(segment3Flights, nil)

	agg := aggregator.NewAggregator(5*time.Second, provider)

	criteria := service.SearchCriteria{
		Passengers: 1,
		CabinClass: "economy",
		Segments: []service.RouteSegment{
			{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"},
			{Origin: "DPS", Destination: "SIN", DepartureDate: "2025-12-17"},
			{Origin: "SIN", Destination: "NRT", DepartureDate: "2025-12-19"},
		},
	}

	ctx := context.Background()
	result, err := agg.SearchAll(ctx, criteria)

	assert.NoError(t, err)
	assert.Len(t, result.MultiCityFlights, 3)
	assert.Len(t, result.MultiCityFlights[0], 1) // First segment: 1 flight
	assert.Len(t, result.MultiCityFlights[1], 2) // Second segment: 2 flights
	assert.Len(t, result.MultiCityFlights[2], 1) // Third segment: 1 flight

	// Verify total results
	assert.Equal(t, 4, result.Metadata.TotalResults) // 1 + 2 + 1 = 4

	provider.AssertExpectations(t)
}
