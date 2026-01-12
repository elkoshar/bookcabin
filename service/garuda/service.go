package garuda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	entity "github.com/elkoshar/bookcabin/service"
)

type Provider struct {
	dataPath string
}

func New(path string) *Provider {
	return &Provider{dataPath: path}
}

func (p *Provider) Name() string { return "Garuda Indonesia" }

func (p *Provider) Search(ctx context.Context, c entity.SearchCriteria) ([]entity.UnifiedFlight, error) {
	content, err := os.ReadFile(p.dataPath)
	if err != nil {
		return nil, fmt.Errorf("garuda read file: %w", err)
	}

	var resp response
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, fmt.Errorf("garuda unmarshal: %w", err)
	}

	var results []entity.UnifiedFlight
	for _, f := range resp.Flights {
		depTime, _ := time.Parse(time.RFC3339, f.Departure.Time)
		arrTime, _ := time.Parse(time.RFC3339, f.Arrival.Time)
		durationMins := int(arrTime.Sub(depTime).Minutes())

		results = append(results, entity.UnifiedFlight{
			ID:             fmt.Sprintf("%s_Garuda", f.FlightID),
			Provider:       p.Name(),
			Airline:        entity.AirlineInfo{Name: f.Airline, Code: "GA"},
			FlightNumber:   f.FlightID,
			Departure:      entity.LocationInfo{Airport: f.Departure.Airport, City: f.Departure.City, DateTime: f.Departure.Time, Timestamp: depTime.Unix()},
			Arrival:        entity.LocationInfo{Airport: f.Arrival.Airport, City: f.Arrival.City, DateTime: f.Arrival.Time, Timestamp: arrTime.Unix()},
			Duration:       entity.DurationInfo{TotalMinutes: durationMins, Formatted: fmt.Sprintf("%dh %dm", durationMins/60, durationMins%60)},
			Stops:          f.Stops,
			Price:          entity.PriceInfo{Amount: f.Price.Amount, Currency: "IDR"},
			AvailableSeats: f.Seats,
			CabinClass:     f.FareClass,
			Amenities:      f.Amenities,
		})
	}
	return results, nil
}
