package aggregator

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/elkoshar/bookcabin/api"
	"github.com/elkoshar/bookcabin/pkg/helpers"
	"github.com/elkoshar/bookcabin/pkg/response"
	"github.com/elkoshar/bookcabin/service"
)

const (
	ErrParseUrlParamMsg = "Parse Url Param Failed. %v"
	ErrCreateDataMsg    = "Create Data Failed. %+v"
	ErrParseValidateMsg = "Failed to Parse and Validate. err=%v"
)

var (
	flightAggregator api.FlightAggregator
)

func Init(service api.FlightAggregator) {
	flightAggregator = service
}

// SearchFlight : HTTP Handler for searching flights
// @Summary Search Flight
// @Description SearchFlight handles request for searching flights
// @Tags Flight
// @Accept json
// @Produce json
// @Param Accept-Language header string true "accept language" default(id)
// @Param body body service.SearchCriteria true "Request Body"
// @Success 200 {object} response.Response{data=service.SearchResponse} "Success Response"
// @Router /flight/search [POST]
func Search(w http.ResponseWriter, r *http.Request) {

	resp := response.Response{}
	defer resp.Render(w, r)

	var (
		err    error
		req    service.SearchCriteria
		result service.SearchResponse
	)

	err = helpers.ParseBodyAndValidate(r, &req)
	if err != nil {
		slog.WarnContext(r.Context(), fmt.Sprintf(ErrParseValidateMsg, err))
		resp.SetError(err, http.StatusBadRequest)
		return
	}

	result, err = flightAggregator.SearchAll(r.Context(), req)
	if err != nil {
		slog.WarnContext(r.Context(), fmt.Sprintf(ErrCreateDataMsg, err))
		resp.SetError(err, http.StatusInternalServerError)
		return
	}

	resp.Data = result
	resp.Code = http.StatusOK

}
