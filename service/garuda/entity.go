package garuda

// entity.go hanya berisi struktur data internal (DTO) untuk parsing JSON

type response struct {
	Status  string   `json:"status"`
	Flights []flight `json:"flights"`
}

type flight struct {
	FlightID  string   `json:"flight_id"`
	Airline   string   `json:"airline"`
	Departure endpoint `json:"departure"`
	Arrival   endpoint `json:"arrival"`
	Price     price    `json:"price"`
	Stops     int      `json:"stops"`
	Seats     int      `json:"available_seats"`
	FareClass string   `json:"fare_class"`
	Amenities []string `json:"amenities"`
}

type endpoint struct {
	Airport string `json:"airport"`
	City    string `json:"city"`
	Time    string `json:"time"`
}

type price struct {
	Amount float64 `json:"amount"`
}
