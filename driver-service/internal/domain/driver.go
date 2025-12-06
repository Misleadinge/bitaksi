package domain

import "time"

// TaxiType represents the type of taxi
type TaxiType string

const (
	TaxiTypeSari    TaxiType = "sari"
	TaxiTypeTurkuaz TaxiType = "turkuaz"
	TaxiTypeSiyah   TaxiType = "siyah"
)

// IsValid checks if the taxi type is valid
func (t TaxiType) IsValid() bool {
	return t == TaxiTypeSari || t == TaxiTypeTurkuaz || t == TaxiTypeSiyah
}

// Location represents geographic coordinates
type Location struct {
	Lat float64 `bson:"lat" json:"lat" example:"41.0431"`
	Lon float64 `bson:"lon" json:"lon" example:"29.0099"`
}

// Driver represents a taxi driver entity
type Driver struct {
	ID        string    `bson:"_id,omitempty" json:"id" example:"507f1f77bcf86cd799439011"`
	FirstName string    `bson:"firstName" json:"firstName" example:"Ahmet"`
	LastName  string    `bson:"lastName" json:"lastName" example:"Demir"`
	Plate     string    `bson:"plate" json:"plate" example:"34ABC123"`
	TaxiType  TaxiType  `bson:"taxiType" json:"taxiType" example:"sari"`
	CarBrand  string    `bson:"carBrand" json:"carBrand" example:"Toyota"`
	CarModel  string    `bson:"carModel" json:"carModel" example:"Corolla"`
	Location  Location  `bson:"location" json:"location"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt" example:"2025-12-06T01:00:00Z"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt" example:"2025-12-06T01:00:00Z"`
}

// DriverRepository defines the interface for driver data access
type DriverRepository interface {
	Create(ctx interface{}, driver *Driver) error
	Update(ctx interface{}, id string, driver *Driver) error
	GetByID(ctx interface{}, id string) (*Driver, error)
	List(ctx interface{}, page, pageSize int) ([]*Driver, int64, error)
	FindNearby(ctx interface{}, lat, lon float64, radiusKm float64, taxiType *TaxiType) ([]*Driver, error)
}
