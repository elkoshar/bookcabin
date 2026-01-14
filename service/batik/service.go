package batik

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
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

func (p *Provider) Name() string { return "Batik Air" }

func (p *Provider) Search(ctx context.Context, c entity.SearchCriteria) ([]entity.UnifiedFlight, error) {
	content, err := os.ReadFile(p.dataPath)
	if err != nil {
		return nil, err
	}

	var resp response
	if err := json.Unmarshal(content, &resp); err != nil {
		return nil, err
	}

	var results []entity.UnifiedFlight
	for _, f := range resp.Results {

		if f.Origin != c.Origin || f.Destination != c.Destination {
			continue
		}

		layout := "2006-01-02T15:04:05-0700"

		locDep := helpers.GetTimezone(f.DepartureDateTime)
		locArr := helpers.GetTimezone(f.ArrivalDateTime)

		depTime, _ := time.ParseInLocation(layout, f.DepartureDateTime, locDep)
		arrTime, _ := time.ParseInLocation(layout, f.ArrivalDateTime, locArr)

		if depTime.Format("2006-01-02") != c.DepartureDate {
			continue
		}
		durationMins := int(arrTime.Sub(depTime).Minutes())

		originCity := helpers.GetCityName(f.Origin)
		destinationCity := helpers.GetCityName(f.Destination)

		results = append(results, entity.UnifiedFlight{
			ID:             fmt.Sprintf("%s_Batik", f.FlightNumber),
			Provider:       p.Name(),
			Airline:        entity.AirlineInfo{Name: f.AirlineName, Code: "ID"},
			FlightNumber:   f.FlightNumber,
			Departure:      entity.LocationInfo{Airport: f.Origin, City: originCity, DateTime: depTime.Format(time.RFC3339), Timestamp: depTime.Unix()},
			Arrival:        entity.LocationInfo{Airport: f.Destination, City: destinationCity, DateTime: arrTime.Format(time.RFC3339), Timestamp: arrTime.Unix()},
			Duration:       entity.DurationInfo{TotalMinutes: durationMins, Formatted: fmt.Sprintf("%dh %dm", durationMins/60, durationMins%60)},
			Price:          entity.PriceInfo{Amount: f.Fare.TotalPrice, Currency: "IDR"},
			AvailableSeats: f.SeatsAvailable,
			CabinClass:     "economy",
		})
	}
	return results, nil
}
