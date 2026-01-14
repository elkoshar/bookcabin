package service

type SearchCriteria struct {
	Origin        string
	Destination   string
	DepartureDate string
	ReturnDate    string
	Passengers    int
	CabinClass    string
}

type UnifiedFlight struct {
	ID             string       `json:"id"`
	Provider       string       `json:"provider"`
	Airline        AirlineInfo  `json:"airline"`
	FlightNumber   string       `json:"flight_number"`
	Departure      LocationInfo `json:"departure"`
	Arrival        LocationInfo `json:"arrival"`
	Duration       DurationInfo `json:"duration"`
	Stops          int          `json:"stops"`
	Price          PriceInfo    `json:"price"`
	AvailableSeats int          `json:"available_seats"`
	CabinClass     string       `json:"cabin_class"`
	Amenities      []string     `json:"amenities"`
	Score          float64      `json:"-"`
}

type AirlineInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

type LocationInfo struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	DateTime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

type DurationInfo struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

type PriceInfo struct {
	Amount    float64 `json:"amount"`
	Currency  string  `json:"currency"`
	Formatted string  `json:"formatted,omitempty"`
}

type SearchResponse struct {
	Criteria SearchCriteria  `json:"search_criteria"`
	Metadata Metadata        `json:"metadata"`
	Flights  []UnifiedFlight `json:"flights"`
}

type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
}
