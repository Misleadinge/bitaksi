package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/bitaksi/driver-service/docs" // swagger docs
	"github.com/bitaksi/driver-service/internal/config"
	"github.com/bitaksi/driver-service/internal/handler"
	"github.com/bitaksi/driver-service/internal/middleware"
	"github.com/bitaksi/driver-service/internal/repository/mongodb"
	"github.com/bitaksi/driver-service/internal/usecase"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.uber.org/zap"
)

// @title Driver Service API
// @version 1.0
// @description TaxiHub Driver Service API
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@bitaksi.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8081
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := initLogger(cfg.Logging.Level)
	defer logger.Sync()

	// Connect to MongoDB
	db, err := connectMongoDB(cfg.MongoDB, logger)
	if err != nil {
		logger.Fatal("failed to connect to MongoDB", zap.Error(err))
	}
	defer func() {
		if err := db.Client().Disconnect(context.Background()); err != nil {
			logger.Error("failed to disconnect from MongoDB", zap.Error(err))
		}
	}()

	// Initialize repository
	driverRepo := mongodb.NewDriverRepository(db, logger)

	// Initialize use case
	driverUseCase := usecase.NewDriverUseCase(driverRepo, logger)

	// Initialize handler
	driverHandler := handler.NewDriverHandler(driverUseCase, logger)

	// Setup router
	router := setupRouter(driverHandler, logger, cfg)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting driver service", zap.String("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start server", zap.Error(err))
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("server forced to shutdown", zap.Error(err))
	}

	logger.Info("server exited")
}

func initLogger(level string) *zap.Logger {
	var zapConfig zap.Config
	if level == "debug" {
		zapConfig = zap.NewDevelopmentConfig()
	} else {
		zapConfig = zap.NewProductionConfig()
	}

	logger, err := zapConfig.Build()
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	return logger
}

func connectMongoDB(cfg config.MongoDBConfig, logger *zap.Logger) (*mongo.Database, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(cfg.URI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping the database
	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	logger.Info("connected to MongoDB", zap.String("database", cfg.Database))
	return client.Database(cfg.Database), nil
}

func setupRouter(driverHandler *handler.DriverHandler, logger *zap.Logger, cfg *config.Config) *gin.Engine {
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Middleware
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler(logger))
	router.Use(middleware.RequestLogger(logger))
	router.Use(gin.Recovery())

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// API routes
	v1 := router.Group("/api/v1")
	{
		drivers := v1.Group("/drivers")
		{
			drivers.POST("", driverHandler.CreateDriver)
			drivers.PUT("/:id", driverHandler.UpdateDriver)
			drivers.GET("/:id", driverHandler.GetDriver)
			drivers.GET("", driverHandler.ListDrivers)
			drivers.GET("/nearby", driverHandler.FindNearbyDrivers)
		}
	}

	// Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return router
}
