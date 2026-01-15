package lion

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

func (p *Provider) Name() string { return "Lion Air" }

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

	for _, f := range resp.Data.Flights {

		if f.Route.From.Code != c.Origin || f.Route.To.Code != c.Destination {
			continue
		}

		if strings.ToLower(f.Pricing.FareType) != strings.ToLower(c.CabinClass) {
			continue
		}

		locDep := helpers.GetTimezone(f.Schedule.Departure)
		locArr := helpers.GetTimezone(f.Schedule.Arrival)

		tDep, _ := time.ParseInLocation("2006-01-02T15:04:05", f.Schedule.Departure, locDep)
		tArr, _ := time.ParseInLocation("2006-01-02T15:04:05", f.Schedule.Arrival, locArr)

		if tDep.Format("2006-01-02") != c.DepartureDate {
			continue
		}

		dur := int(tArr.Sub(tDep).Minutes())

		results = append(results, entity.UnifiedFlight{
			ID:             fmt.Sprintf("%s_Lion", f.ID),
			Provider:       p.Name(),
			Airline:        entity.AirlineInfo{Name: f.Carrier.Name, Code: f.Carrier.Iata},
			FlightNumber:   f.ID,
			Departure:      entity.LocationInfo{Airport: f.Route.From.Code, City: f.Route.From.City, DateTime: tDep.Format(time.RFC3339), Timestamp: tDep.Unix()},
			Arrival:        entity.LocationInfo{Airport: f.Route.To.Code, City: f.Route.To.City, DateTime: tArr.Format(time.RFC3339), Timestamp: tArr.Unix()},
			Duration:       entity.DurationInfo{TotalMinutes: dur, Formatted: fmt.Sprintf("%dh %dm", dur/60, dur%60)},
			Price:          entity.PriceInfo{Amount: f.Pricing.Total, Currency: "IDR"},
			AvailableSeats: f.SeatsLeft,
			CabinClass:     "economy",
		})
	}
	return results, nil
}
