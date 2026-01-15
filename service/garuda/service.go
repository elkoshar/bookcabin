package garuda

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/elkoshar/bookcabin/pkg/helpers"
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

		if f.Departure.Airport != c.Origin || f.Arrival.Airport != c.Destination {
			continue
		}

		if strings.ToLower(f.FareClass) != strings.ToLower(c.CabinClass) {
			continue
		}

		locDep := helpers.GetTimezone(f.Departure.Time)
		locArr := helpers.GetTimezone(f.Arrival.Time)

		depTime, _ := time.ParseInLocation(time.RFC3339, f.Departure.Time, locDep)
		arrTime, _ := time.ParseInLocation(time.RFC3339, f.Arrival.Time, locArr)

		if depTime.Format("2006-01-02") != c.DepartureDate {
			continue
		}
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
