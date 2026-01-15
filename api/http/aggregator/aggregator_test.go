package aggregator_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/elkoshar/bookcabin/api"
	"github.com/elkoshar/bookcabin/api/http/aggregator"
	"github.com/elkoshar/bookcabin/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockFlightAggregator implements api.FlightAggregator for testing
type MockFlightAggregator struct {
	mock.Mock
}

var _ api.FlightAggregator = (*MockFlightAggregator)(nil)

func (m *MockFlightAggregator) SearchAll(ctx context.Context, criteria service.SearchCriteria) (service.SearchResponse, error) {
	args := m.Called(ctx, criteria)
	return args.Get(0).(service.SearchResponse), args.Error(1)
}

func TestInit(t *testing.T) {
	mockService := &MockFlightAggregator{}

	// Test that Init doesn't panic and accepts the service
	assert.NotPanics(t, func() {
		aggregator.Init(mockService)
	})
}

func TestSearch_Success(t *testing.T) {
	// Setup mock service
	mockService := &MockFlightAggregator{}
	aggregator.Init(mockService)

	// Mock response
	expectedResponse := service.SearchResponse{
		Criteria: service.SearchCriteria{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
		},
		Flights: []service.UnifiedFlight{
			{
				ID:           "TEST123",
				Provider:     "Test Provider",
				FlightNumber: "TP100",
				Price:        service.PriceInfo{Amount: 1500000, Currency: "IDR"},
				Duration:     service.DurationInfo{TotalMinutes: 120},
			},
		},
		Metadata: service.Metadata{
			TotalResults:       1,
			ProvidersQueried:   1,
			ProvidersSucceeded: 1,
			ProvidersFailed:    0,
			SearchTimeMs:       50,
		},
	}

	mockService.On("SearchAll", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "CGK" && criteria.Destination == "DPS"
	})).Return(expectedResponse, nil)

	// Create request
	requestBody := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/flight/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	aggregator.Search(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Equal(t, float64(200), response["code"])
	assert.NotNil(t, response["data"])

	// Verify data content
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["flights"])
	assert.NotNil(t, data["metadata"])

	flights := data["flights"].([]interface{})
	assert.Len(t, flights, 1)

	metadata := data["metadata"].(map[string]interface{})
	assert.Equal(t, float64(1), metadata["total_results"])
	assert.Equal(t, float64(1), metadata["providers_succeeded"])

	mockService.AssertExpectations(t)
}

func TestSearch_InvalidJSON(t *testing.T) {
	mockService := &MockFlightAggregator{}
	aggregator.Init(mockService)

	// Create request with invalid JSON
	req := httptest.NewRequest(http.MethodPost, "/flight/search", bytes.NewBuffer([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	aggregator.Search(w, req)

	// Check response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify error response
	assert.Equal(t, float64(400), response["code"])
	assert.NotNil(t, response["error"])

	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, true, errorObj["status"])
}

func TestSearch_ValidationError(t *testing.T) {
	mockService := &MockFlightAggregator{}
	aggregator.Init(mockService)

	// Create request with invalid data (missing required fields)
	requestBody := service.SearchCriteria{
		// Missing Origin and Destination
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	// Mock service to return error for invalid criteria
	mockService.On("SearchAll", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.Origin == "" || criteria.Destination == ""
	})).Return(service.SearchResponse{}, errors.New("invalid search criteria"))

	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/flight/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	aggregator.Search(w, req)

	// Check response - could be either validation error (400) or service error (500)
	// depending on whether validation catches the issue or service does
	assert.True(t, w.Code == http.StatusBadRequest || w.Code == http.StatusInternalServerError)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify error response
	assert.NotNil(t, response["error"])

	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, true, errorObj["status"])

	if w.Code == http.StatusInternalServerError {
		mockService.AssertExpectations(t)
	}
}

func TestSearch_ServiceError(t *testing.T) {
	// Setup mock service
	mockService := &MockFlightAggregator{}
	aggregator.Init(mockService)

	expectedError := errors.New("service unavailable")

	// Mock service to return error
	mockService.On("SearchAll", mock.Anything, mock.Anything).Return(service.SearchResponse{}, expectedError)

	// Create valid request
	requestBody := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/flight/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	aggregator.Search(w, req)

	// Check response
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify error response
	assert.Equal(t, float64(500), response["code"])
	assert.NotNil(t, response["error"])

	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, true, errorObj["status"])

	mockService.AssertExpectations(t)
}

func TestSearch_EmptyBody(t *testing.T) {
	mockService := &MockFlightAggregator{}
	aggregator.Init(mockService)

	// Create request with empty body
	req := httptest.NewRequest(http.MethodPost, "/flight/search", bytes.NewBuffer([]byte("")))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	aggregator.Search(w, req)

	// Check response - should be bad request due to empty body
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Parse response body
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify error response
	assert.Equal(t, float64(400), response["code"])
	assert.NotNil(t, response["error"])
}

func TestSearch_RoundTripSearch(t *testing.T) {
	// Setup mock service
	mockService := &MockFlightAggregator{}
	aggregator.Init(mockService)

	// Mock response for round trip
	expectedResponse := service.SearchResponse{
		Criteria: service.SearchCriteria{
			Origin:        "CGK",
			Destination:   "DPS",
			DepartureDate: "2025-12-15",
			ReturnDate:    "2025-12-20",
		},
		Flights: []service.UnifiedFlight{
			{
				ID:           "DEPART123",
				Provider:     "Test Provider",
				FlightNumber: "TP100",
				Price:        service.PriceInfo{Amount: 1500000, Currency: "IDR"},
			},
		},
		ReturnFlights: []service.UnifiedFlight{
			{
				ID:           "RETURN123",
				Provider:     "Test Provider",
				FlightNumber: "TP200",
				Price:        service.PriceInfo{Amount: 1600000, Currency: "IDR"},
			},
		},
		Metadata: service.Metadata{
			TotalResults:       2,
			ProvidersQueried:   2,
			ProvidersSucceeded: 2,
			ProvidersFailed:    0,
		},
	}

	mockService.On("SearchAll", mock.Anything, mock.MatchedBy(func(criteria service.SearchCriteria) bool {
		return criteria.ReturnDate != ""
	})).Return(expectedResponse, nil)

	// Create round trip request
	requestBody := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		ReturnDate:    "2025-12-20",
		Passengers:    1,
		CabinClass:    "economy",
	}

	jsonBody, err := json.Marshal(requestBody)
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/flight/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Language", "en")

	// Create response recorder
	w := httptest.NewRecorder()

	// Call handler
	aggregator.Search(w, req)

	// Check response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse response body
	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify response structure
	assert.Equal(t, float64(200), response["code"])

	// Should have both departure and return flights
	data := response["data"].(map[string]interface{})
	assert.NotNil(t, data["flights"])
	assert.NotNil(t, data["return_flights"])

	returnFlights := data["return_flights"].([]interface{})
	assert.Len(t, returnFlights, 1)

	mockService.AssertExpectations(t)
}
