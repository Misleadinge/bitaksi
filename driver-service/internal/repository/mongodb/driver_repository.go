package mongodb

import (
	"context"
	"errors"
	"time"

	"github.com/bitaksi/driver-service/internal/domain"
	"github.com/bitaksi/driver-service/pkg/haversine"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// DriverRepository implements domain.DriverRepository using MongoDB
type DriverRepository struct {
	collection *mongo.Collection
	logger     *zap.Logger
}

// NewDriverRepository creates a new MongoDB driver repository
func NewDriverRepository(db *mongo.Database, logger *zap.Logger) *DriverRepository {
	return &DriverRepository{
		collection: db.Collection("drivers"),
		logger:     logger,
	}
}

// Create inserts a new driver into MongoDB
func (r *DriverRepository) Create(ctx interface{}, driver *domain.Driver) error {
	c, ok := ctx.(context.Context)
	if !ok {
		c = context.Background()
	}

	driver.CreatedAt = time.Now()
	driver.UpdatedAt = time.Now()

	result, err := r.collection.InsertOne(c, driver)
	if err != nil {
		r.logger.Error("failed to create driver", zap.Error(err))
		return err
	}

	if oid, ok := result.InsertedID.(primitive.ObjectID); ok {
		driver.ID = oid.Hex()
	}

	return nil
}

// Update updates an existing driver in MongoDB
func (r *DriverRepository) Update(ctx interface{}, id string, driver *domain.Driver) error {
	c, ok := ctx.(context.Context)
	if !ok {
		c = context.Background()
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid driver ID")
	}

	driver.UpdatedAt = time.Now()

	filter := bson.M{"_id": objectID}
	update := bson.M{
		"$set": bson.M{
			"firstName": driver.FirstName,
			"lastName":  driver.LastName,
			"plate":     driver.Plate,
			"taxiType":  driver.TaxiType,
			"carBrand":  driver.CarBrand,
			"carModel":  driver.CarModel,
			"location":  driver.Location,
			"updatedAt": driver.UpdatedAt,
		},
	}

	result, err := r.collection.UpdateOne(c, filter, update)
	if err != nil {
		r.logger.Error("failed to update driver", zap.Error(err), zap.String("id", id))
		return err
	}

	if result.MatchedCount == 0 {
		return errors.New("driver not found")
	}

	return nil
}

// GetByID retrieves a driver by ID
func (r *DriverRepository) GetByID(ctx interface{}, id string) (*domain.Driver, error) {
	c, ok := ctx.(context.Context)
	if !ok {
		c = context.Background()
	}

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid driver ID")
	}

	var driver domain.Driver
	filter := bson.M{"_id": objectID}

	err = r.collection.FindOne(c, filter).Decode(&driver)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, errors.New("driver not found")
		}
		r.logger.Error("failed to get driver by ID", zap.Error(err), zap.String("id", id))
		return nil, err
	}

	driver.ID = objectID.Hex()
	return &driver, nil
}

// List retrieves a paginated list of drivers
func (r *DriverRepository) List(ctx interface{}, page, pageSize int) ([]*domain.Driver, int64, error) {
	c, ok := ctx.(context.Context)
	if !ok {
		c = context.Background()
	}

	skip := (page - 1) * pageSize

	// Get total count
	totalCount, err := r.collection.CountDocuments(c, bson.M{})
	if err != nil {
		r.logger.Error("failed to count drivers", zap.Error(err))
		return nil, 0, err
	}

	// Get paginated results
	findOptions := options.Find()
	findOptions.SetSkip(int64(skip))
	findOptions.SetLimit(int64(pageSize))
	findOptions.SetSort(bson.M{"createdAt": -1})

	cursor, err := r.collection.Find(c, bson.M{}, findOptions)
	if err != nil {
		r.logger.Error("failed to list drivers", zap.Error(err))
		return nil, 0, err
	}
	defer cursor.Close(c)

	var driversData []struct {
		ID        primitive.ObjectID `bson:"_id"`
		FirstName string             `bson:"firstName"`
		LastName  string             `bson:"lastName"`
		Plate     string             `bson:"plate"`
		TaxiType  domain.TaxiType    `bson:"taxiType"`
		CarBrand  string             `bson:"carBrand"`
		CarModel  string             `bson:"carModel"`
		Location  domain.Location    `bson:"location"`
		CreatedAt time.Time          `bson:"createdAt"`
		UpdatedAt time.Time          `bson:"updatedAt"`
	}

	if err = cursor.All(c, &driversData); err != nil {
		r.logger.Error("failed to decode drivers", zap.Error(err))
		return nil, 0, err
	}

	// Convert to domain.Driver with string ID
	drivers := make([]*domain.Driver, len(driversData))
	for i, d := range driversData {
		drivers[i] = &domain.Driver{
			ID:        d.ID.Hex(),
			FirstName: d.FirstName,
			LastName:  d.LastName,
			Plate:     d.Plate,
			TaxiType:  d.TaxiType,
			CarBrand:  d.CarBrand,
			CarModel:  d.CarModel,
			Location:  d.Location,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		}
	}

	return drivers, totalCount, nil
}

// FindNearby finds drivers within a specified radius
func (r *DriverRepository) FindNearby(ctx interface{}, lat, lon float64, radiusKm float64, taxiType *domain.TaxiType) ([]*domain.Driver, error) {
	c, ok := ctx.(context.Context)
	if !ok {
		c = context.Background()
	}

	// Build filter
	filter := bson.M{}

	// Add taxi type filter if provided
	if taxiType != nil {
		filter["taxiType"] = *taxiType
	}

	// Get all drivers (we'll filter by distance in memory since MongoDB geospatial queries
	// require a geospatial index and we want to use Haversine formula)
	cursor, err := r.collection.Find(c, filter)
	if err != nil {
		r.logger.Error("failed to find nearby drivers", zap.Error(err))
		return nil, err
	}
	defer cursor.Close(c)

	var allDrivers []struct {
		ID        primitive.ObjectID `bson:"_id"`
		FirstName string             `bson:"firstName"`
		LastName  string             `bson:"lastName"`
		Plate     string             `bson:"plate"`
		TaxiType  domain.TaxiType    `bson:"taxiType"`
		CarBrand  string             `bson:"carBrand"`
		CarModel  string             `bson:"carModel"`
		Location  domain.Location    `bson:"location"`
		CreatedAt time.Time          `bson:"createdAt"`
		UpdatedAt time.Time          `bson:"updatedAt"`
	}

	if err = cursor.All(c, &allDrivers); err != nil {
		r.logger.Error("failed to decode drivers", zap.Error(err))
		return nil, err
	}

	// Filter by distance using Haversine formula and sort by distance
	type driverWithDistance struct {
		driver   *domain.Driver
		distance float64
	}

	var nearbyDrivers []driverWithDistance
	for _, d := range allDrivers {
		distance := haversine.Distance(lat, lon, d.Location.Lat, d.Location.Lon)
		if distance <= radiusKm {
			driver := &domain.Driver{
				ID:        d.ID.Hex(),
				FirstName: d.FirstName,
				LastName:  d.LastName,
				Plate:     d.Plate,
				TaxiType:  d.TaxiType,
				CarBrand:  d.CarBrand,
				CarModel:  d.CarModel,
				Location:  d.Location,
				CreatedAt: d.CreatedAt,
				UpdatedAt: d.UpdatedAt,
			}
			nearbyDrivers = append(nearbyDrivers, driverWithDistance{
				driver:   driver,
				distance: distance,
			})
		}
	}

	// Sort by distance (nearest first) - simple bubble sort
	for i := 0; i < len(nearbyDrivers)-1; i++ {
		for j := i + 1; j < len(nearbyDrivers); j++ {
			if nearbyDrivers[i].distance > nearbyDrivers[j].distance {
				nearbyDrivers[i], nearbyDrivers[j] = nearbyDrivers[j], nearbyDrivers[i]
			}
		}
	}

	// Return only drivers
	result := make([]*domain.Driver, len(nearbyDrivers))
	for i, nd := range nearbyDrivers {
		result[i] = nd.driver
	}

	return result, nil
}
