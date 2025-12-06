package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/bitaksi/driver-service/internal/domain"
	"go.uber.org/zap"
)

// mockDriverRepository is a mock implementation of DriverRepository
type mockDriverRepository struct {
	drivers map[string]*domain.Driver
}

func newMockDriverRepository() *mockDriverRepository {
	return &mockDriverRepository{
		drivers: make(map[string]*domain.Driver),
	}
}

func (m *mockDriverRepository) Create(ctx interface{}, driver *domain.Driver) error {
	if driver.Plate == "" {
		return errors.New("plate is required")
	}
	// Generate unique ID if not set
	if driver.ID == "" {
		driver.ID = "test-id-" + driver.Plate
	}
	m.drivers[driver.ID] = driver
	return nil
}

func (m *mockDriverRepository) Update(ctx interface{}, id string, driver *domain.Driver) error {
	if _, exists := m.drivers[id]; !exists {
		return errors.New("driver not found")
	}
	m.drivers[id] = driver
	return nil
}

func (m *mockDriverRepository) GetByID(ctx interface{}, id string) (*domain.Driver, error) {
	driver, exists := m.drivers[id]
	if !exists {
		return nil, errors.New("driver not found")
	}
	return driver, nil
}

func (m *mockDriverRepository) List(ctx interface{}, page, pageSize int) ([]*domain.Driver, int64, error) {
	drivers := make([]*domain.Driver, 0, len(m.drivers))
	for _, driver := range m.drivers {
		drivers = append(drivers, driver)
	}
	return drivers, int64(len(drivers)), nil
}

func (m *mockDriverRepository) FindNearby(ctx interface{}, lat, lon float64, radiusKm float64, taxiType *domain.TaxiType) ([]*domain.Driver, error) {
	drivers := make([]*domain.Driver, 0)
	for _, driver := range m.drivers {
		if taxiType == nil || driver.TaxiType == *taxiType {
			drivers = append(drivers, driver)
		}
	}
	return drivers, nil
}

func TestDriverUseCase_CreateDriver(t *testing.T) {
	logger := zap.NewNop()
	repo := newMockDriverRepository()
	uc := NewDriverUseCase(repo, logger)

	tests := []struct {
		name    string
		req     *CreateDriverRequest
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid request",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: false,
		},
		{
			name: "missing first name",
			req: &CreateDriverRequest{
				LastName: "Demir",
				Plate:    "34ABC123",
				TaxiType: domain.TaxiTypeSari,
				CarBrand: "Toyota",
				CarModel: "Corolla",
				Lat:      41.0431,
				Lon:      29.0099,
			},
			wantErr: true,
			errMsg:  "firstName is required",
		},
		{
			name: "invalid taxi type",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiType("invalid"),
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: true,
			errMsg:  "invalid taxiType",
		},
		{
			name: "invalid plate format",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "INVALID",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: true,
			errMsg:  "plate must be in format",
		},
		{
			name: "invalid latitude",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       100.0, // Invalid
				Lon:       29.0099,
			},
			wantErr: true,
			errMsg:  "latitude must be between",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := uc.CreateDriver(context.Background(), tt.req)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
					return
				}
				if tt.errMsg != "" && err.Error() != tt.errMsg && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error message containing %q, got %q", tt.errMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
					return
				}
				if driver == nil {
					t.Errorf("expected driver but got nil")
				}
			}
		})
	}
}

func TestDriverUseCase_UpdateDriver(t *testing.T) {
	logger := zap.NewNop()
	repo := newMockDriverRepository()
	uc := NewDriverUseCase(repo, logger)

	// Create a driver first
	createReq := &CreateDriverRequest{
		FirstName: "Ahmet",
		LastName:  "Demir",
		Plate:     "34ABC123",
		TaxiType:  domain.TaxiTypeSari,
		CarBrand:  "Toyota",
		CarModel:  "Corolla",
		Lat:       41.0431,
		Lon:       29.0099,
	}
	driver, _ := uc.CreateDriver(context.Background(), createReq)
	driverID := driver.ID

	tests := []struct {
		name    string
		id      string
		req     *UpdateDriverRequest
		wantErr bool
	}{
		{
			name: "update first name",
			id:   driverID,
			req: &UpdateDriverRequest{
				FirstName: stringPtr("Mehmet"),
			},
			wantErr: false,
		},
		{
			name: "update location",
			id:   driverID,
			req: &UpdateDriverRequest{
				Lat: float64Ptr(41.0082),
				Lon: float64Ptr(28.9784),
			},
			wantErr: false,
		},
		{
			name: "driver not found",
			id:   "non-existent-id",
			req: &UpdateDriverRequest{
				FirstName: stringPtr("Mehmet"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := uc.UpdateDriver(context.Background(), tt.id, tt.req)
			if tt.wantErr && err == nil {
				t.Errorf("expected error but got none")
			} else if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestDriverUseCase_ListDrivers(t *testing.T) {
	logger := zap.NewNop()
	repo := newMockDriverRepository()
	uc := NewDriverUseCase(repo, logger)

	// Create some drivers
	for i := 0; i < 5; i++ {
		req := &CreateDriverRequest{
			FirstName: "Driver",
			LastName:  "Test",
			Plate:     "34ABC" + string(rune('0'+i)),
			TaxiType:  domain.TaxiTypeSari,
			CarBrand:  "Toyota",
			CarModel:  "Corolla",
			Lat:       41.0431,
			Lon:       29.0099,
		}
		uc.CreateDriver(context.Background(), req)
	}

	response, err := uc.ListDrivers(context.Background(), 1, 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if response.TotalCount != 5 {
		t.Errorf("expected total count 5, got %d", response.TotalCount)
	}

	if len(response.Drivers) != 5 {
		t.Errorf("expected 5 drivers, got %d", len(response.Drivers))
	}
}

func stringPtr(s string) *string {
	return &s
}

func float64Ptr(f float64) *float64 {
	return &f
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr ||
			s[len(s)-len(substr):] == substr ||
			containsMiddle(s, substr))))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
