package aggregator

import (
	"net/http"

	"github.com/elkoshar/bookcabin/pkg/response"
	"github.com/elkoshar/bookcabin/service/aggregator"
)

var (
	flightAggregator aggregator.FlightAggregator
)

func Init(service aggregator.FlightAggregator) {
	flightAggregator = service
}

// SearchFlight : HTTP Handler for searching flights
// @Summary Search Flight
// @Description SearchFlight handles request for searching flights
// @Tags Flight
// @Accept json
// @Produce json
// @Param Accept-Language header string true "accept language" default(id)
// @Router /flight/search [POST]
func SearchFlight(w http.ResponseWriter, r *http.Request) {
	resp := response.Response{}
	defer resp.Render(w, r)

}
