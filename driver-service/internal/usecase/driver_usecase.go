package usecase

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/bitaksi/driver-service/internal/domain"
	"github.com/bitaksi/driver-service/pkg/haversine"
	"go.uber.org/zap"
)

// DriverUseCase defines the interface for driver business logic
type DriverUseCase interface {
	CreateDriver(ctx context.Context, req *CreateDriverRequest) (*domain.Driver, error)
	UpdateDriver(ctx context.Context, id string, req *UpdateDriverRequest) (*domain.Driver, error)
	GetDriver(ctx context.Context, id string) (*domain.Driver, error)
	ListDrivers(ctx context.Context, page, pageSize int) (*ListDriversResponse, error)
	FindNearbyDrivers(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*NearbyDriverResponse, error)
}

// CreateDriverRequest represents the request to create a driver
type CreateDriverRequest struct {
	FirstName string          `json:"firstName" example:"Ahmet" binding:"required"`
	LastName  string          `json:"lastName" example:"Demir" binding:"required"`
	Plate     string          `json:"plate" example:"34ABC123" binding:"required"`
	TaxiType  domain.TaxiType `json:"taksiType" example:"sari" binding:"required"`
	CarBrand  string          `json:"carBrand" example:"Toyota" binding:"required"`
	CarModel  string          `json:"carModel" example:"Corolla" binding:"required"`
	Lat       float64         `json:"lat" example:"41.0431" binding:"required"`
	Lon       float64         `json:"lon" example:"29.0099" binding:"required"`
}

// UpdateDriverRequest represents the request to update a driver
type UpdateDriverRequest struct {
	FirstName *string          `json:"firstName,omitempty" example:"Mehmet"`
	LastName  *string          `json:"lastName,omitempty" example:"Kurt"`
	Plate     *string          `json:"plate,omitempty" example:"34XYZ789"`
	TaxiType  *domain.TaxiType `json:"taksiType,omitempty" example:"turkuaz"`
	CarBrand  *string          `json:"carBrand,omitempty" example:"Honda"`
	CarModel  *string          `json:"carModel,omitempty" example:"Civic"`
	Location  *domain.Location `json:"location,omitempty"` // Nested location object
}

// ListDriversResponse represents the paginated list response
type ListDriversResponse struct {
	Drivers    []*domain.Driver `json:"drivers"`
	TotalCount int64            `json:"totalCount" example:"1"`
	Page       int              `json:"page" example:"1"`
	PageSize   int              `json:"pageSize" example:"20"`
}

// NearbyDriverResponse represents a driver in nearby search results
type NearbyDriverResponse struct {
	ID         string  `json:"id" example:"507f1f77bcf86cd799439011"`
	FirstName  string  `json:"firstName" example:"Ahmet"`
	LastName   string  `json:"lastName" example:"Demir"`
	Plate      string  `json:"plate" example:"34ABC123"`
	TaxiType   string  `json:"taxiType" example:"sari"`
	DistanceKm float64 `json:"distanceKm" example:"0.5"`
}

// driverUseCase implements DriverUseCase
type driverUseCase struct {
	repo   domain.DriverRepository
	logger *zap.Logger
}

// NewDriverUseCase creates a new driver use case
func NewDriverUseCase(repo domain.DriverRepository, logger *zap.Logger) DriverUseCase {
	return &driverUseCase{
		repo:   repo,
		logger: logger,
	}
}

// CreateDriver creates a new driver
func (uc *driverUseCase) CreateDriver(ctx context.Context, req *CreateDriverRequest) (*domain.Driver, error) {
	// Validate input
	if err := uc.validateCreateRequest(req); err != nil {
		return nil, err
	}

	driver := &domain.Driver{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Plate:     strings.ToUpper(req.Plate),
		TaxiType:  req.TaxiType,
		CarBrand:  req.CarBrand,
		CarModel:  req.CarModel,
		Location: domain.Location{
			Lat: req.Lat,
			Lon: req.Lon,
		},
	}

	if err := uc.repo.Create(ctx, driver); err != nil {
		uc.logger.Error("failed to create driver", zap.Error(err))
		return nil, errors.New("failed to create driver")
	}

	uc.logger.Info("driver created", zap.String("id", driver.ID), zap.String("plate", driver.Plate))
	return driver, nil
}

// UpdateDriver updates an existing driver
func (uc *driverUseCase) UpdateDriver(ctx context.Context, id string, req *UpdateDriverRequest) (*domain.Driver, error) {
	// Get existing driver
	existing, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("driver not found")
	}

	// Update fields if provided
	if req.FirstName != nil {
		if *req.FirstName == "" {
			return nil, errors.New("firstName cannot be empty")
		}
		existing.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		if *req.LastName == "" {
			return nil, errors.New("lastName cannot be empty")
		}
		existing.LastName = *req.LastName
	}
	if req.Plate != nil {
		if err := uc.validatePlate(*req.Plate); err != nil {
			return nil, err
		}
		existing.Plate = strings.ToUpper(*req.Plate)
	}
	if req.TaxiType != nil {
		if !req.TaxiType.IsValid() {
			return nil, fmt.Errorf("invalid taxiType: %s", *req.TaxiType)
		}
		existing.TaxiType = *req.TaxiType
	}
	if req.CarBrand != nil {
		if *req.CarBrand == "" {
			return nil, errors.New("carBrand cannot be empty")
		}
		existing.CarBrand = *req.CarBrand
	}
	if req.CarModel != nil {
		if *req.CarModel == "" {
			return nil, errors.New("carModel cannot be empty")
		}
		existing.CarModel = *req.CarModel
	}
	// Update location if provided
	if req.Location != nil {
		if err := uc.validateLocation(req.Location.Lat, req.Location.Lon); err != nil {
			return nil, err
		}
		existing.Location.Lat = req.Location.Lat
		existing.Location.Lon = req.Location.Lon
	}

	if err := uc.repo.Update(ctx, id, existing); err != nil {
		uc.logger.Error("failed to update driver", zap.Error(err), zap.String("id", id))
		return nil, errors.New("failed to update driver")
	}

	uc.logger.Info("driver updated", zap.String("id", id))
	return existing, nil
}

// GetDriver retrieves a driver by ID
func (uc *driverUseCase) GetDriver(ctx context.Context, id string) (*domain.Driver, error) {
	driver, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.New("driver not found")
	}
	return driver, nil
}

// ListDrivers retrieves a paginated list of drivers
func (uc *driverUseCase) ListDrivers(ctx context.Context, page, pageSize int) (*ListDriversResponse, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	drivers, totalCount, err := uc.repo.List(ctx, page, pageSize)
	if err != nil {
		uc.logger.Error("failed to list drivers", zap.Error(err))
		return nil, errors.New("failed to list drivers")
	}

	return &ListDriversResponse{
		Drivers:    drivers,
		TotalCount: totalCount,
		Page:       page,
		PageSize:   pageSize,
	}, nil
}

// FindNearbyDrivers finds drivers within 6km radius
func (uc *driverUseCase) FindNearbyDrivers(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*NearbyDriverResponse, error) {
	// Validate location
	if err := uc.validateLocation(lat, lon); err != nil {
		return nil, err
	}

	// Validate taxi type if provided
	if taxiType != nil && !taxiType.IsValid() {
		return nil, fmt.Errorf("invalid taxiType: %s", *taxiType)
	}

	const radiusKm = 6.0
	drivers, err := uc.repo.FindNearby(ctx, lat, lon, radiusKm, taxiType)
	if err != nil {
		uc.logger.Error("failed to find nearby drivers", zap.Error(err))
		return nil, errors.New("failed to find nearby drivers")
	}

	// Convert to response format with distance
	responses := make([]*NearbyDriverResponse, len(drivers))
	for i, driver := range drivers {
		// Calculate distance for response
		// Note: We already filtered by distance, but we need to recalculate for the response
		// In a real implementation, we might want to store the distance in the repository
		// For now, we'll use a simple approach and recalculate
		distance := haversine.Distance(lat, lon, driver.Location.Lat, driver.Location.Lon)

		responses[i] = &NearbyDriverResponse{
			ID:         driver.ID,
			FirstName:  driver.FirstName,
			LastName:   driver.LastName,
			Plate:      driver.Plate,
			TaxiType:   string(driver.TaxiType),
			DistanceKm: distance,
		}
	}

	uc.logger.Info("found nearby drivers", zap.Int("count", len(responses)))
	return responses, nil
}

// validateCreateRequest validates the create driver request
func (uc *driverUseCase) validateCreateRequest(req *CreateDriverRequest) error {
	if req.FirstName == "" {
		return errors.New("firstName is required")
	}
	if req.LastName == "" {
		return errors.New("lastName is required")
	}
	if err := uc.validatePlate(req.Plate); err != nil {
		return err
	}
	if !req.TaxiType.IsValid() {
		return fmt.Errorf("invalid taxiType: %s. Must be one of: sari, turkuaz, siyah", req.TaxiType)
	}
	if req.CarBrand == "" {
		return errors.New("carBrand is required")
	}
	if req.CarModel == "" {
		return errors.New("carModel is required")
	}
	if err := uc.validateLocation(req.Lat, req.Lon); err != nil {
		return err
	}
	return nil
}

// validatePlate validates Turkish license plate format (simplified: 2-3 digits + 1-3 letters + 1-4 digits)
func (uc *driverUseCase) validatePlate(plate string) error {
	if plate == "" {
		return errors.New("plate is required")
	}
	// Turkish plate format: 34ABC123 or 34AB123 or 34A123
	plateRegex := regexp.MustCompile(`^[0-9]{2,3}[A-Z]{1,3}[0-9]{1,4}$`)
	if !plateRegex.MatchString(strings.ToUpper(plate)) {
		return errors.New("plate must be in format: 2-3 digits, 1-3 letters, 1-4 digits (e.g., 34ABC123)")
	}
	return nil
}

// validateLocation validates latitude and longitude
func (uc *driverUseCase) validateLocation(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return errors.New("latitude must be between -90 and 90")
	}
	if lon < -180 || lon > 180 {
		return errors.New("longitude must be between -180 and 180")
	}
	return nil
}
