package airasia_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/elkoshar/bookcabin/service"
	"github.com/elkoshar/bookcabin/service/airasia"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	path := "/path/to/data.json"
	provider := airasia.New(path)

	assert.NotNil(t, provider)
}

func TestProvider_Name(t *testing.T) {
	provider := airasia.New("")

	assert.Equal(t, "AirAsia", provider.Name())
}

func TestProvider_Search_Success(t *testing.T) {
	// Create temporary test data
	testData := map[string]interface{}{
		"flights": []map[string]interface{}{
			{
				"flight_code":   "QZ520",
				"airline":       "AirAsia",
				"from_airport":  "CGK",
				"to_airport":    "DPS",
				"depart_time":   "2025-12-15T04:45:00+07:00",
				"arrive_time":   "2025-12-15T07:25:00+08:00",
				"price_idr":     1500000.0,
				"direct_flight": true,
				"seats":         150,
				"cabin_class":   "economy",
			},
			{
				"flight_code":   "QZ521",
				"airline":       "AirAsia",
				"from_airport":  "CGK",
				"to_airport":    "DPS",
				"depart_time":   "2025-12-15T14:30:00+07:00",
				"arrive_time":   "2025-12-15T17:15:00+08:00",
				"price_idr":     1800000.0,
				"direct_flight": false,
				"seats":         120,
				"cabin_class":   "economy",
			},
		},
	}

	// Create temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := airasia.New(tmpFile)

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
	assert.Len(t, flights, 2)

	// Check first flight
	flight1 := flights[0]
	assert.Equal(t, "QZ520_AirAsia", flight1.ID)
	assert.Equal(t, "AirAsia", flight1.Provider)
	assert.Equal(t, "AirAsia", flight1.Airline.Name)
	assert.Equal(t, "QZ", flight1.Airline.Code)
	assert.Equal(t, "QZ520", flight1.FlightNumber)
	assert.Equal(t, "CGK", flight1.Departure.Airport)
	assert.Equal(t, "DPS", flight1.Arrival.Airport)
	assert.Equal(t, 0, flight1.Stops)
	assert.Equal(t, float64(1500000), flight1.Price.Amount)
	assert.Equal(t, "IDR", flight1.Price.Currency)
	assert.Equal(t, 150, flight1.AvailableSeats)
	assert.Equal(t, "economy", flight1.CabinClass)

	// Check second flight (with stop)
	flight2 := flights[1]
	assert.Equal(t, "QZ521_AirAsia", flight2.ID)
	assert.Equal(t, 1, flight2.Stops)
	assert.Equal(t, float64(1800000), flight2.Price.Amount)
}

func TestProvider_Search_NoMatchingFlights(t *testing.T) {
	testData := map[string]interface{}{
		"flights": []map[string]interface{}{
			{
				"flight_code":   "QZ520",
				"airline":       "AirAsia",
				"from_airport":  "CGK",
				"to_airport":    "DPS",
				"depart_time":   "2025-12-16T04:45:00+07:00", // Different date
				"arrive_time":   "2025-12-16T07:25:00+08:00",
				"price_idr":     1500000.0,
				"direct_flight": true,
				"seats":         150,
				"cabin_class":   "economy",
			},
		},
	}

	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test_data.json")

	data, err := json.Marshal(testData)
	assert.NoError(t, err)

	err = os.WriteFile(tmpFile, data, 0644)
	assert.NoError(t, err)

	provider := airasia.New(tmpFile)

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
	provider := airasia.New("/nonexistent/file.json")

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

	provider := airasia.New(tmpFile)

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
