package airasia

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

func (p *Provider) Name() string { return "AirAsia" }

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
	for _, f := range resp.Flights {

		if f.FromAirport != c.Origin || f.ToAirport != c.Destination {
			continue
		}

		locDep := helpers.GetTimezone(f.DepartTime)
		locArr := helpers.GetTimezone(f.ArriveTime)

		depTime, _ := time.ParseInLocation(time.RFC3339, f.DepartTime, locDep)
		arrTime, _ := time.ParseInLocation(time.RFC3339, f.ArriveTime, locArr)

		if depTime.Format("2006-01-02") != c.DepartureDate {
			continue
		}

		durationMins := int(arrTime.Sub(depTime).Minutes())

		stops := 0
		if !f.DirectFlight {
			stops = 1
		}
		originCity := helpers.GetCityName(f.FromAirport)
		destinationCity := helpers.GetCityName(f.ToAirport)

		results = append(results, entity.UnifiedFlight{
			ID:             fmt.Sprintf("%s_AirAsia", f.FlightCode),
			Provider:       p.Name(),
			Airline:        entity.AirlineInfo{Name: f.Airline, Code: "QZ"},
			FlightNumber:   f.FlightCode,
			Departure:      entity.LocationInfo{Airport: f.FromAirport, City: originCity, DateTime: f.DepartTime, Timestamp: depTime.Unix()},
			Arrival:        entity.LocationInfo{Airport: f.ToAirport, City: destinationCity, DateTime: f.ArriveTime, Timestamp: arrTime.Unix()},
			Duration:       entity.DurationInfo{TotalMinutes: durationMins, Formatted: fmt.Sprintf("%dh %dm", durationMins/60, durationMins%60)},
			Stops:          stops,
			Price:          entity.PriceInfo{Amount: f.PriceIDR, Currency: "IDR"},
			AvailableSeats: f.Seats,
			CabinClass:     f.CabinClass,
		})
	}
	return results, nil
}
