package handler

// CreateDriverRequest represents the request to create a driver
type CreateDriverRequest struct {
	FirstName string  `json:"firstName" example:"Ahmet" binding:"required"`
	LastName  string  `json:"lastName" example:"Demir" binding:"required"`
	Plate     string  `json:"plate" example:"34ABC123" binding:"required"`
	TaxiType  string  `json:"taksiType" example:"sari" enums:"sari,turkuaz,siyah" binding:"required"`
	CarBrand  string  `json:"carBrand" example:"Toyota" binding:"required"`
	CarModel  string  `json:"carModel" example:"Corolla" binding:"required"`
	Lat       float64 `json:"lat" example:"41.0431" binding:"required"`
	Lon       float64 `json:"lon" example:"29.0099" binding:"required"`
}

// UpdateDriverRequest represents the request to update a driver
type UpdateDriverRequest struct {
	FirstName *string  `json:"firstName,omitempty" example:"Mehmet"`
	LastName  *string  `json:"lastName,omitempty" example:"Kurt"`
	Plate     *string  `json:"plate,omitempty" example:"34XYZ789"`
	TaxiType  *string  `json:"taksiType,omitempty" example:"turkuaz" enums:"sari,turkuaz,siyah"`
	CarBrand  *string  `json:"carBrand,omitempty" example:"Honda"`
	CarModel  *string  `json:"carModel,omitempty" example:"Civic"`
	Lat       *float64 `json:"lat,omitempty" example:"41.0082"`
	Lon       *float64 `json:"lon,omitempty" example:"28.9784"`
}
