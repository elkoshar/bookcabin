package lion_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/elkoshar/bookcabin/service"
	"github.com/elkoshar/bookcabin/service/lion"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	path := "/path/to/data.json"
	provider := lion.New(path)

	assert.NotNil(t, provider)
}

func TestProvider_Name(t *testing.T) {
	provider := lion.New("")

	assert.Equal(t, "Lion Air", provider.Name())
}

func TestProvider_Search_Success(t *testing.T) {
	testData := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"available_flights": []map[string]interface{}{
				{
					"id": "JT610",
					"carrier": map[string]interface{}{
						"name": "Lion Air",
						"iata": "JT",
					},
					"route": map[string]interface{}{
						"from": map[string]interface{}{
							"code": "CGK",
							"name": "Soekarno-Hatta International",
							"city": "Jakarta",
						},
						"to": map[string]interface{}{
							"code": "DPS",
							"name": "Ngurah Rai International",
							"city": "Denpasar",
						},
					},
					"schedule": map[string]interface{}{
						"departure": "2025-12-15T05:30:00",
						"arrival":   "2025-12-15T08:45:00",
					},
					"pricing": map[string]interface{}{
						"total": 1250000.0,
					},
					"seats_left": 180,
				},
			},
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := lion.New(tmpFile)

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, criteria)

	assert.NoError(t, err)
	assert.Len(t, flights, 1)

	flight := flights[0]
	assert.Equal(t, "JT610_Lion", flight.ID)
	assert.Equal(t, "Lion Air", flight.Provider)
	assert.Equal(t, "Lion Air", flight.Airline.Name)
	assert.Equal(t, "JT", flight.Airline.Code)
	assert.Equal(t, "JT610", flight.FlightNumber)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, 0, flight.Stops)
	assert.Equal(t, float64(1250000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 180, flight.AvailableSeats)
	assert.Equal(t, "economy", flight.CabinClass)
}

func TestProvider_Search_NoMatchingFlights(t *testing.T) {
	testData := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"available_flights": []map[string]interface{}{
				{
					"id": "JT610",
					"carrier": map[string]interface{}{
						"name": "Lion Air",
						"iata": "JT",
					},
					"route": map[string]interface{}{
						"from": map[string]interface{}{
							"code": "CGK",
							"name": "Soekarno-Hatta International",
							"city": "Jakarta",
						},
						"to": map[string]interface{}{
							"code": "DPS",
							"name": "Ngurah Rai International",
							"city": "Denpasar",
						},
					},
					"schedule": map[string]interface{}{
						"departure": "2025-12-16T05:30:00", // Different date
						"arrival":   "2025-12-16T08:45:00",
					},
					"pricing": map[string]interface{}{
						"total": 1250000.0,
					},
					"seats_left": 180,
				},
			},
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := lion.New(tmpFile)

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15", // Different date
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, criteria)

	assert.NoError(t, err)
	assert.Empty(t, flights)
}

func TestProvider_Search_FileNotFound(t *testing.T) {
	provider := lion.New("/nonexistent/file.json")

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, criteria)

	assert.Error(t, err)
	assert.Nil(t, flights)
}

func TestProvider_Search_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(tmpFile, []byte("invalid json"), 0644)
	assert.NoError(t, err)

	provider := lion.New(tmpFile)

	criteria := service.SearchCriteria{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	ctx := context.Background()
	flights, err := provider.Search(ctx, criteria)

	assert.Error(t, err)
	assert.Nil(t, flights)
}
