package lion

type response struct {
	Data data `json:"data"`
}

type data struct {
	Flights []flight `json:"available_flights"`
}

type flight struct {
	ID        string   `json:"id"`
	Carrier   carrier  `json:"carrier"`
	Route     route    `json:"route"`
	Schedule  schedule `json:"schedule"`
	Pricing   pricing  `json:"pricing"`
	SeatsLeft int      `json:"seats_left"`
}

type carrier struct {
	Name string `json:"name"`
	Iata string `json:"iata"`
}

type route struct {
	From location `json:"from"`
	To   location `json:"to"`
}

type location struct {
	Code string `json:"code"`
	City string `json:"city"`
}

type schedule struct {
	Departure string `json:"departure"`
	Arrival   string `json:"arrival"`
}

type pricing struct {
	Total float64 `json:"total"`
}
