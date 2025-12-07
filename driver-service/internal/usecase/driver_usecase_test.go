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
	drivers              map[string]*domain.Driver
	shouldFailCreate     bool
	shouldFailUpdate     bool
	shouldFailList       bool
	shouldFailGet        bool
	shouldFailFindNearby bool
}

func newMockDriverRepository() *mockDriverRepository {
	return &mockDriverRepository{
		drivers: make(map[string]*domain.Driver),
	}
}

func (m *mockDriverRepository) Create(ctx interface{}, driver *domain.Driver) error {
	if m.shouldFailCreate {
		return errors.New("repository error")
	}
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
	if m.shouldFailUpdate {
		return errors.New("repository error")
	}
	if _, exists := m.drivers[id]; !exists {
		return errors.New("driver not found")
	}
	m.drivers[id] = driver
	return nil
}

func (m *mockDriverRepository) GetByID(ctx interface{}, id string) (*domain.Driver, error) {
	if m.shouldFailGet {
		return nil, errors.New("repository error")
	}
	driver, exists := m.drivers[id]
	if !exists {
		return nil, errors.New("driver not found")
	}
	return driver, nil
}

func (m *mockDriverRepository) List(ctx interface{}, page, pageSize int) ([]*domain.Driver, int64, error) {
	if m.shouldFailList {
		return nil, 0, errors.New("repository error")
	}
	drivers := make([]*domain.Driver, 0, len(m.drivers))
	for _, driver := range m.drivers {
		drivers = append(drivers, driver)
	}
	// Simulate pagination
	start := (page - 1) * pageSize
	if start < 0 {
		start = 0
	}
	end := start + pageSize
	if end > len(drivers) {
		end = len(drivers)
	}
	if start >= len(drivers) {
		return []*domain.Driver{}, int64(len(drivers)), nil
	}
	return drivers[start:end], int64(len(drivers)), nil
}

func (m *mockDriverRepository) FindNearby(ctx interface{}, lat, lon float64, radiusKm float64, taxiType *domain.TaxiType) ([]*domain.Driver, error) {
	if m.shouldFailFindNearby {
		return nil, errors.New("repository error")
	}
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
			name: "plate with lowercase letters",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34abc123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: false, // Should be converted to uppercase
		},
		{
			name: "valid plate format 2 digits",
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
			name: "valid plate format 3 digits",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "345ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: false,
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
		{
			name: "missing last name",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: true,
			errMsg:  "lastName is required",
		},
		{
			name: "missing plate",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: true,
			errMsg:  "plate is required",
		},
		{
			name: "missing car brand",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: true,
			errMsg:  "carBrand is required",
		},
		{
			name: "missing car model",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				Lat:       41.0431,
				Lon:       29.0099,
			},
			wantErr: true,
			errMsg:  "carModel is required",
		},
		{
			name: "invalid longitude",
			req: &CreateDriverRequest{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Lat:       41.0431,
				Lon:       200.0, // Invalid
			},
			wantErr: true,
			errMsg:  "longitude must be between",
		},
		{
			name: "repository error on create",
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
			wantErr: true,
			errMsg:  "failed to create driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockDriverRepository()
			if tt.name == "repository error on create" {
				repo.shouldFailCreate = true
			}
			uc := NewDriverUseCase(repo, logger)
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
			name: "update last name",
			id:   driverID,
			req: &UpdateDriverRequest{
				LastName: stringPtr("Kurt"),
			},
			wantErr: false,
		},
		{
			name: "update plate",
			id:   driverID,
			req: &UpdateDriverRequest{
				Plate: stringPtr("34XYZ789"),
			},
			wantErr: false,
		},
		{
			name: "update taxi type",
			id:   driverID,
			req: &UpdateDriverRequest{
				TaxiType: func() *domain.TaxiType { t := domain.TaxiTypeTurkuaz; return &t }(),
			},
			wantErr: false,
		},
		{
			name: "update car brand",
			id:   driverID,
			req: &UpdateDriverRequest{
				CarBrand: stringPtr("Honda"),
			},
			wantErr: false,
		},
		{
			name: "update car model",
			id:   driverID,
			req: &UpdateDriverRequest{
				CarModel: stringPtr("Civic"),
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
			name: "update location with only lat",
			id:   driverID,
			req: &UpdateDriverRequest{
				Lat: float64Ptr(41.0082),
			},
			wantErr: true, // Both lat and lon must be provided
		},
		{
			name: "update location with only lon",
			id:   driverID,
			req: &UpdateDriverRequest{
				Lon: float64Ptr(28.9784),
			},
			wantErr: true, // Both lat and lon must be provided
		},
		{
			name: "update with empty first name",
			id:   driverID,
			req: &UpdateDriverRequest{
				FirstName: stringPtr(""),
			},
			wantErr: true,
		},
		{
			name: "update with empty last name",
			id:   driverID,
			req: &UpdateDriverRequest{
				LastName: stringPtr(""),
			},
			wantErr: true,
		},
		{
			name: "update with invalid plate",
			id:   driverID,
			req: &UpdateDriverRequest{
				Plate: stringPtr("INVALID"),
			},
			wantErr: true,
		},
		{
			name: "update with invalid taxi type",
			id:   driverID,
			req: &UpdateDriverRequest{
				TaxiType: func() *domain.TaxiType { t := domain.TaxiType("invalid"); return &t }(),
			},
			wantErr: true,
		},
		{
			name: "update with invalid location",
			id:   driverID,
			req: &UpdateDriverRequest{
				Lat: float64Ptr(100.0),
				Lon: float64Ptr(28.9784),
			},
			wantErr: true,
		},
		{
			name: "update with empty car brand",
			id:   driverID,
			req: &UpdateDriverRequest{
				CarBrand: stringPtr(""),
			},
			wantErr: true,
		},
		{
			name: "update with empty car model",
			id:   driverID,
			req: &UpdateDriverRequest{
				CarModel: stringPtr(""),
			},
			wantErr: true,
		},
		{
			name: "update with repository error",
			id:   driverID,
			req: &UpdateDriverRequest{
				FirstName: stringPtr("Mehmet"),
			},
			wantErr: true,
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
			repo := newMockDriverRepository()
			uc := NewDriverUseCase(repo, logger)

			// Create a driver first for update tests
			if tt.name != "driver not found" {
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
				if tt.name == "update with repository error" {
					repo.shouldFailUpdate = true
					tt.id = driver.ID
				} else if tt.id == driverID {
					tt.id = driver.ID
				}
			}

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

	tests := []struct {
		name     string
		page     int
		pageSize int
		wantErr  bool
	}{
		{
			name:     "first page",
			page:     1,
			pageSize: 10,
			wantErr:  false,
		},
		{
			name:     "pagination",
			page:     1,
			pageSize: 2,
			wantErr:  false,
		},
		{
			name:     "empty page",
			page:     10,
			pageSize: 10,
			wantErr:  false,
		},
		{
			name:     "page less than 1",
			page:     0,
			pageSize: 10,
			wantErr:  false, // Should default to 1
		},
		{
			name:     "pageSize less than 1",
			page:     1,
			pageSize: 0,
			wantErr:  false, // Should default to 20
		},
		{
			name:     "pageSize greater than 100",
			page:     1,
			pageSize: 200,
			wantErr:  false, // Should cap at 100
		},
		{
			name:     "repository error on list",
			page:     1,
			pageSize: 10,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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

			if tt.name == "repository error on list" {
				repo.shouldFailList = true
			}

			response, err := uc.ListDrivers(context.Background(), tt.page, tt.pageSize)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if response == nil {
					t.Errorf("expected response but got nil")
				}
			}
		})
	}
}

func TestDriverUseCase_GetDriver(t *testing.T) {
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
		wantErr bool
	}{
		{
			name:    "valid driver",
			id:      driverID,
			wantErr: false,
		},
		{
			name:    "driver not found",
			id:      "non-existent-id",
			wantErr: true,
		},
		{
			name:    "empty id",
			id:      "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := uc.GetDriver(context.Background(), tt.id)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if driver == nil {
					t.Errorf("expected driver but got nil")
				}
				if driver != nil && driver.ID != tt.id {
					t.Errorf("expected driver ID %s, got %s", tt.id, driver.ID)
				}
			}
		})
	}
}

func TestDriverUseCase_FindNearbyDrivers(t *testing.T) {
	logger := zap.NewNop()
	repo := newMockDriverRepository()
	uc := NewDriverUseCase(repo, logger)

	// Create drivers at different locations
	locations := []struct {
		lat, lon float64
		plate    string
	}{
		{41.0431, 29.0099, "34ABC123"}, // Close to search point
		{41.0082, 28.9784, "34XYZ789"}, // Close to search point
		{39.9334, 32.8597, "06DEF456"}, // Far (Ankara)
	}

	for _, loc := range locations {
		req := &CreateDriverRequest{
			FirstName: "Driver",
			LastName:  "Test",
			Plate:     loc.plate,
			TaxiType:  domain.TaxiTypeSari,
			CarBrand:  "Toyota",
			CarModel:  "Corolla",
			Lat:       loc.lat,
			Lon:       loc.lon,
		}
		uc.CreateDriver(context.Background(), req)
	}

	tests := []struct {
		name      string
		lat       float64
		lon       float64
		taxiType  *domain.TaxiType
		wantErr   bool
		wantCount int
	}{
		{
			name:      "find nearby without filter",
			lat:       41.0431,
			lon:       29.0099,
			taxiType:  nil,
			wantErr:   false,
			wantCount: 3, // All drivers (mock returns all)
		},
		{
			name:      "find nearby with taxi type filter",
			lat:       41.0431,
			lon:       29.0099,
			taxiType:  func() *domain.TaxiType { t := domain.TaxiTypeSari; return &t }(),
			wantErr:   false,
			wantCount: 3,
		},
		{
			name:     "invalid latitude",
			lat:      100.0,
			lon:      29.0099,
			taxiType: nil,
			wantErr:  true,
		},
		{
			name:     "invalid longitude",
			lat:      41.0431,
			lon:      200.0,
			taxiType: nil,
			wantErr:  true,
		},
		{
			name:     "invalid taxi type",
			lat:      41.0431,
			lon:      29.0099,
			taxiType: func() *domain.TaxiType { t := domain.TaxiType("invalid"); return &t }(),
			wantErr:  true,
		},
		{
			name:      "repository error",
			lat:       41.0431,
			lon:       29.0099,
			taxiType:  nil,
			wantErr:   true,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := newMockDriverRepository()
			uc := NewDriverUseCase(repo, logger)

			// Create drivers at different locations
			if tt.name != "repository error" {
				locations := []struct {
					lat, lon float64
					plate    string
				}{
					{41.0431, 29.0099, "34ABC123"},
					{41.0082, 28.9784, "34XYZ789"},
					{39.9334, 32.8597, "06DEF456"},
				}

				for _, loc := range locations {
					req := &CreateDriverRequest{
						FirstName: "Driver",
						LastName:  "Test",
						Plate:     loc.plate,
						TaxiType:  domain.TaxiTypeSari,
						CarBrand:  "Toyota",
						CarModel:  "Corolla",
						Lat:       loc.lat,
						Lon:       loc.lon,
					}
					uc.CreateDriver(context.Background(), req)
				}
			}

			if tt.name == "repository error" {
				repo.shouldFailFindNearby = true
			}

			drivers, err := uc.FindNearbyDrivers(context.Background(), tt.lat, tt.lon, tt.taxiType)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if drivers == nil {
					t.Errorf("expected drivers but got nil")
				}
				if tt.wantCount > 0 && len(drivers) != tt.wantCount {
					t.Errorf("expected %d drivers, got %d", tt.wantCount, len(drivers))
				}
			}
		})
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
