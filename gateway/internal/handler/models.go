package handler

// Driver represents a taxi driver
type Driver struct {
	ID        string `json:"id"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Plate     string `json:"plate"`
	TaxiType  string `json:"taxiType"`
	CarBrand  string `json:"carBrand"`
	CarModel  string `json:"carModel"`
	Location  struct {
		Lat float64 `json:"lat"`
		Lon float64 `json:"lon"`
	} `json:"location"`
	CreatedAt string `json:"createdAt"`
	UpdatedAt string `json:"updatedAt"`
}

// ListDriversResponse represents a paginated list of drivers
type ListDriversResponse struct {
	Drivers    []Driver `json:"drivers"`
	TotalCount int64    `json:"totalCount"`
	Page       int      `json:"page"`
	PageSize   int      `json:"pageSize"`
}

// NearbyDriverResponse represents a driver in nearby search results
type NearbyDriverResponse struct {
	ID         string  `json:"id"`
	FirstName  string  `json:"firstName"`
	LastName   string  `json:"lastName"`
	Plate      string  `json:"plate"`
	TaxiType   string  `json:"taxiType"`
	DistanceKm float64 `json:"distanceKm"`
}
