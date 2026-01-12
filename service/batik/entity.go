package batik

type response struct {
	Results []result `json:"results"`
}

type result struct {
	FlightNumber      string `json:"flightNumber"`
	AirlineName       string `json:"airlineName"`
	Origin            string `json:"origin"`
	Destination       string `json:"destination"`
	DepartureDateTime string `json:"departureDateTime"` // Format: 2025-12-15T07:15:00+0700
	ArrivalDateTime   string `json:"arrivalDateTime"`
	Fare              fare   `json:"fare"`
	SeatsAvailable    int    `json:"seatsAvailable"`
}

type fare struct {
	TotalPrice float64 `json:"totalPrice"`
	Class      string  `json:"class"`
}
