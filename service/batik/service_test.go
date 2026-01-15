package batik_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/elkoshar/bookcabin/service"
	"github.com/elkoshar/bookcabin/service/batik"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	path := "/path/to/data.json"
	provider := batik.New(path)

	assert.NotNil(t, provider)
}

func TestProvider_Name(t *testing.T) {
	provider := batik.New("")

	assert.Equal(t, "Batik Air", provider.Name())
}

func TestProvider_Search_Success(t *testing.T) {
	testData := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"flightNumber":      "ID6420",
				"airlineName":       "Batik Air",
				"origin":            "CGK",
				"destination":       "DPS",
				"departureDateTime": "2025-12-15T07:15:00+0700",
				"arrivalDateTime":   "2025-12-15T10:30:00+0800",
				"fare": map[string]interface{}{
					"totalPrice": 2100000.0,
					"class":      "economy",
				},
				"seatsAvailable": 89,
			},
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := batik.New(tmpFile)

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
	assert.Equal(t, "ID6420_Batik", flight.ID)
	assert.Equal(t, "Batik Air", flight.Provider)
	assert.Equal(t, "Batik Air", flight.Airline.Name)
	assert.Equal(t, "ID", flight.Airline.Code)
	assert.Equal(t, "ID6420", flight.FlightNumber)
	assert.Equal(t, "CGK", flight.Departure.Airport)
	assert.Equal(t, "DPS", flight.Arrival.Airport)
	assert.Equal(t, 0, flight.Stops)
	assert.Equal(t, float64(2100000), flight.Price.Amount)
	assert.Equal(t, "IDR", flight.Price.Currency)
	assert.Equal(t, 89, flight.AvailableSeats)
	assert.Equal(t, "economy", flight.CabinClass)
}

func TestProvider_Search_NoMatchingFlights(t *testing.T) {
	testData := map[string]interface{}{
		"results": []map[string]interface{}{
			{
				"flightNumber":      "ID6420",
				"airlineName":       "Batik Air",
				"origin":            "CGK",
				"destination":       "DPS",
				"departureDateTime": "2025-12-16T07:15:00+0700", // Different date
				"arrivalDateTime":   "2025-12-16T10:30:00+0800",
				"fare": map[string]interface{}{
					"totalPrice": 2100000.0,
					"class":      "economy",
				},
				"seatsAvailable": 89,
			},
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := batik.New(tmpFile)

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
	provider := batik.New("/nonexistent/file.json")

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

	provider := batik.New(tmpFile)

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
