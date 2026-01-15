package garuda_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/elkoshar/bookcabin/service"
	"github.com/elkoshar/bookcabin/service/garuda"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	path := "/path/to/data.json"
	provider := garuda.New(path)

	assert.NotNil(t, provider)
}

func TestProvider_Name(t *testing.T) {
	provider := garuda.New("")

	assert.Equal(t, "Garuda Indonesia", provider.Name())
}

func TestProvider_Search_Success(t *testing.T) {
	testData := map[string]interface{}{
		"status": "success",
		"flights": []map[string]interface{}{
			{
				"flight_id": "GA100",
				"airline":   "Garuda Indonesia",
				"departure": map[string]interface{}{
					"airport": "CGK",
					"city":    "Jakarta",
					"time":    "2025-12-15T06:00:00+07:00",
				},
				"arrival": map[string]interface{}{
					"airport": "DPS",
					"city":    "Denpasar",
					"time":    "2025-12-15T09:30:00+08:00",
				},
				"price": map[string]interface{}{
					"amount":   3500000.0,
					"currency": "IDR",
				},
				"stops":           0,
				"available_seats": 75,
				"fare_class":      "economy",
			},
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := garuda.New(tmpFile)

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
	assert.Equal(t, "GA100_Garuda", flight.ID)
	assert.Equal(t, "Garuda Indonesia", flight.Provider)
	assert.Equal(t, "Garuda Indonesia", flight.Airline.Name)
	assert.Equal(t, "GA", flight.Airline.Code)
	assert.Equal(t, "GA100", flight.FlightNumber)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, 0, flight.Stops)
	assert.Equal(t, float64(3500000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 75, flight.AvailableSeats)
	assert.Equal(t, "economy", flight.CabinClass)
}

func TestProvider_Search_NoMatchingFlights(t *testing.T) {
	testData := map[string]interface{}{
		"status": "success",
		"flights": []map[string]interface{}{
			{
				"flight_id": "GA100",
				"airline":   "Garuda Indonesia",
				"departure": map[string]interface{}{
					"airport": "CGK",
					"city":    "Jakarta",
					"time":    "2025-12-16T06:00:00+07:00", // Different date
				},
				"arrival": map[string]interface{}{
					"airport": "DPS",
					"city":    "Denpasar",
					"time":    "2025-12-16T09:30:00+08:00",
				},
				"price": map[string]interface{}{
					"amount":   3500000.0,
					"currency": "IDR",
				},
				"stops":           0,
				"available_seats": 75,
				"fare_class":      "economy",
			},
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := garuda.New(tmpFile)

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
	provider := garuda.New("/nonexistent/file.json")

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
	assert.Contains(t, err.Error(), "garuda read file")
}

func TestProvider_Search_InvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "invalid.json")

	err := os.WriteFile(tmpFile, []byte("invalid json"), 0644)
	assert.NoError(t, err)

	provider := garuda.New(tmpFile)

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
