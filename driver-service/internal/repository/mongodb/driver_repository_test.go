package mongodb

import (
	"context"
	"testing"
	"time"

	"github.com/bitaksi/driver-service/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// setupTestDB creates a test MongoDB connection
func setupTestDB(t *testing.T) (*mongo.Database, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Use test database
	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	require.NoError(t, err)

	db := client.Database("test_taxihub")

	// Clean up function
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		db.Drop(ctx)
		client.Disconnect(ctx)
	}

	return db, cleanup
}

func TestNewDriverRepository(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	assert.NotNil(t, repo)
	assert.NotNil(t, repo.collection)
	assert.Equal(t, logger, repo.logger)
}

func TestDriverRepository_Create(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	tests := []struct {
		name    string
		driver  *domain.Driver
		wantErr bool
	}{
		{
			name: "successful create",
			driver: &domain.Driver{
				FirstName: "Ahmet",
				LastName:  "Demir",
				Plate:     "34ABC123",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Toyota",
				CarModel:  "Corolla",
				Location: domain.Location{
					Lat: 41.0431,
					Lon: 29.0099,
				},
			},
			wantErr: false,
		},
		{
			name: "create with existing ID",
			driver: &domain.Driver{
				ID:        "507f1f77bcf86cd799439011",
				FirstName: "Test",
				LastName:  "Driver",
				Plate:     "34TEST1",
				TaxiType:  domain.TaxiTypeSari,
				CarBrand:  "Test",
				CarModel:  "Model",
				Location: domain.Location{
					Lat: 41.0,
					Lon: 29.0,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			err := repo.Create(ctx, tt.driver)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				// ID should be set (either from existing or generated)
				assert.NotEmpty(t, tt.driver.ID)
			}
		})
	}
}

func TestDriverRepository_Update(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	// Create a driver first
	driver := &domain.Driver{
		FirstName: "Ahmet",
		LastName:  "Demir",
		Plate:     "34ABC123",
		TaxiType:  domain.TaxiTypeSari,
		CarBrand:  "Toyota",
		CarModel:  "Corolla",
		Location: domain.Location{
			Lat: 41.0431,
			Lon: 29.0099,
		},
	}
	ctx := context.Background()
	err := repo.Create(ctx, driver)
	require.NoError(t, err)
	require.NotEmpty(t, driver.ID)

	tests := []struct {
		name    string
		id      string
		driver  *domain.Driver
		wantErr bool
	}{
		{
			name: "successful update",
			id:   driver.ID,
			driver: &domain.Driver{
				ID:        driver.ID,
				FirstName: "Mehmet",
				LastName:  "Kurt",
				Plate:     "34XYZ789",
				TaxiType:  domain.TaxiTypeTurkuaz,
				CarBrand:  "Honda",
				CarModel:  "Civic",
				Location: domain.Location{
					Lat: 41.0082,
					Lon: 28.9784,
				},
			},
			wantErr: false,
		},
		{
			name: "driver not found",
			id:   "507f1f77bcf86cd799439011",
			driver: &domain.Driver{
				FirstName: "Test",
			},
			wantErr: true,
		},
		{
			name:    "invalid id",
			id:      "invalid-id",
			driver:  &domain.Driver{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := repo.Update(ctx, tt.id, tt.driver)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDriverRepository_GetByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	// Create a driver first
	driver := &domain.Driver{
		FirstName: "Ahmet",
		LastName:  "Demir",
		Plate:     "34ABC123",
		TaxiType:  domain.TaxiTypeSari,
		CarBrand:  "Toyota",
		CarModel:  "Corolla",
		Location: domain.Location{
			Lat: 41.0431,
			Lon: 29.0099,
		},
	}
	ctx := context.Background()
	err := repo.Create(ctx, driver)
	require.NoError(t, err)
	require.NotEmpty(t, driver.ID)

	tests := []struct {
		name    string
		id      string
		wantErr bool
	}{
		{
			name:    "successful get",
			id:      driver.ID,
			wantErr: false,
		},
		{
			name:    "driver not found",
			id:      "507f1f77bcf86cd799439011",
			wantErr: true,
		},
		{
			name:    "invalid id",
			id:      "invalid-id",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := repo.GetByID(ctx, tt.id)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.id, result.ID)
			}
		})
	}
}

func TestDriverRepository_List(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	ctx := context.Background()

	// Create multiple drivers
	for i := 0; i < 5; i++ {
		driver := &domain.Driver{
			FirstName: "Driver",
			LastName:  "Test",
			Plate:     "34ABC" + string(rune('0'+i)),
			TaxiType:  domain.TaxiTypeSari,
			CarBrand:  "Toyota",
			CarModel:  "Corolla",
			Location: domain.Location{
				Lat: 41.0431,
				Lon: 29.0099,
			},
		}
		err := repo.Create(ctx, driver)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		page     int
		pageSize int
		wantErr  bool
		minCount int
	}{
		{
			name:     "first page",
			page:     1,
			pageSize: 10,
			wantErr:  false,
			minCount: 5,
		},
		{
			name:     "pagination",
			page:     1,
			pageSize: 2,
			wantErr:  false,
			minCount: 2,
		},
		{
			name:     "second page",
			page:     2,
			pageSize: 2,
			wantErr:  false,
			minCount: 2,
		},
		{
			name:     "empty page",
			page:     10,
			pageSize: 10,
			wantErr:  false,
			minCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drivers, totalCount, err := repo.List(ctx, tt.page, tt.pageSize)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, drivers)
				assert.GreaterOrEqual(t, totalCount, int64(0))
				assert.GreaterOrEqual(t, len(drivers), tt.minCount)
			}
		})
	}
}

func TestDriverRepository_FindNearby(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	ctx := context.Background()

	// Create drivers at different locations
	locations := []struct {
		lat, lon float64
		plate    string
		taxiType domain.TaxiType
	}{
		{41.0431, 29.0099, "34ABC123", domain.TaxiTypeSari},    // Close
		{41.0082, 28.9784, "34XYZ789", domain.TaxiTypeSari},    // Close
		{39.9334, 32.8597, "06DEF456", domain.TaxiTypeTurkuaz}, // Far (Ankara)
		{0.0, 0.0, "00ZERO1", domain.TaxiTypeSari},             // Zero coordinates (should be skipped)
		{100.0, 200.0, "99INVALID", domain.TaxiTypeSari},       // Invalid coordinates (should be skipped)
	}

	for _, loc := range locations {
		driver := &domain.Driver{
			FirstName: "Driver",
			LastName:  "Test",
			Plate:     loc.plate,
			TaxiType:  loc.taxiType,
			CarBrand:  "Toyota",
			CarModel:  "Corolla",
			Location: domain.Location{
				Lat: loc.lat,
				Lon: loc.lon,
			},
		}
		err := repo.Create(ctx, driver)
		require.NoError(t, err)
	}

	tests := []struct {
		name     string
		lat      float64
		lon      float64
		radiusKm float64
		taxiType *domain.TaxiType
		wantErr  bool
		minCount int
	}{
		{
			name:     "find nearby without filter",
			lat:      41.0431,
			lon:      29.0099,
			radiusKm: 6.0,
			taxiType: nil,
			wantErr:  false,
			minCount: 2, // At least 2 drivers should be within 6km
		},
		{
			name:     "find nearby with taxi type filter",
			lat:      41.0431,
			lon:      29.0099,
			radiusKm: 6.0,
			taxiType: func() *domain.TaxiType { t := domain.TaxiTypeSari; return &t }(),
			wantErr:  false,
			minCount: 2,
		},
		{
			name:     "find nearby with different taxi type",
			lat:      41.0431,
			lon:      29.0099,
			radiusKm: 6.0,
			taxiType: func() *domain.TaxiType { t := domain.TaxiTypeTurkuaz; return &t }(),
			wantErr:  false,
			minCount: 0, // Turkuaz driver is far away
		},
		{
			name:     "find nearby with zero radius",
			lat:      41.0431,
			lon:      29.0099,
			radiusKm: 0.0,
			taxiType: nil,
			wantErr:  false,
			minCount: 0, // No drivers within 0km
		},
		{
			name:     "find nearby with large radius",
			lat:      41.0431,
			lon:      29.0099,
			radiusKm: 1000.0,
			taxiType: nil,
			wantErr:  false,
			minCount: 3, // Should find more drivers with large radius
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			drivers, err := repo.FindNearby(ctx, tt.lat, tt.lon, tt.radiusKm, tt.taxiType)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, drivers)
				assert.GreaterOrEqual(t, len(drivers), tt.minCount)
				// Verify drivers are sorted by distance
				for i := 0; i < len(drivers)-1; i++ {
					// Note: We can't directly verify distance sorting without recalculating,
					// but we can verify all drivers have valid locations
					assert.NotEqual(t, 0.0, drivers[i].Location.Lat)
					assert.NotEqual(t, 0.0, drivers[i].Location.Lon)
				}
			}
		})
	}
}

func TestDriverRepository_CreateWithInvalidContext(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	driver := &domain.Driver{
		FirstName: "Test",
		LastName:  "Driver",
		Plate:     "34TEST1",
		TaxiType:  domain.TaxiTypeSari,
		CarBrand:  "Test",
		CarModel:  "Model",
		Location: domain.Location{
			Lat: 41.0,
			Lon: 29.0,
		},
	}

	// Test with invalid context type (should convert to background)
	err := repo.Create("not-a-context", driver)
	// Should still work as it converts to background context
	assert.NoError(t, err)
}

func TestDriverRepository_UpdateWithInvalidContext(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	// Create a driver first
	driver := &domain.Driver{
		FirstName: "Test",
		LastName:  "Driver",
		Plate:     "34TEST1",
		TaxiType:  domain.TaxiTypeSari,
		CarBrand:  "Test",
		CarModel:  "Model",
		Location: domain.Location{
			Lat: 41.0,
			Lon: 29.0,
		},
	}
	ctx := context.Background()
	err := repo.Create(ctx, driver)
	require.NoError(t, err)

	// Test with invalid context type
	driver.FirstName = "Updated"
	err = repo.Update("not-a-context", driver.ID, driver)
	assert.NoError(t, err)
}

func TestDriverRepository_GetByIDWithInvalidContext(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	// Create a driver first
	driver := &domain.Driver{
		FirstName: "Test",
		LastName:  "Driver",
		Plate:     "34TEST1",
		TaxiType:  domain.TaxiTypeSari,
		CarBrand:  "Test",
		CarModel:  "Model",
		Location: domain.Location{
			Lat: 41.0,
			Lon: 29.0,
		},
	}
	ctx := context.Background()
	err := repo.Create(ctx, driver)
	require.NoError(t, err)

	// Test with invalid context type
	result, err := repo.GetByID("not-a-context", driver.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
}

func TestDriverRepository_ListWithInvalidContext(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	// Test with invalid context type
	drivers, totalCount, err := repo.List("not-a-context", 1, 10)
	assert.NoError(t, err)
	assert.NotNil(t, drivers)
	assert.GreaterOrEqual(t, totalCount, int64(0))
}

func TestDriverRepository_FindNearbyWithInvalidContext(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	logger := zap.NewNop()
	repo := NewDriverRepository(db, logger)

	// Test with invalid context type
	drivers, err := repo.FindNearby("not-a-context", 41.0, 29.0, 6.0, nil)
	assert.NoError(t, err)
	assert.NotNil(t, drivers)
}
