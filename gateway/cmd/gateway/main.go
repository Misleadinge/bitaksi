package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/bitaksi/gateway/docs" // swagger docs
	"github.com/bitaksi/gateway/internal/config"
	"github.com/bitaksi/gateway/internal/handler"
	"github.com/bitaksi/gateway/internal/middleware"
	"github.com/bitaksi/gateway/internal/service"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"go.uber.org/zap"
)

// @title Gateway API
// @version 1.0
// @description TaxiHub API Gateway
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.email support@bitaksi.com

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Enter your JWT token with "Bearer " prefix (e.g., "Bearer eyJhbGci..."). IMPORTANT: You must include "Bearer " before the token.

// @host localhost:8080
// @BasePath /
func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize logger
	logger := initLogger(cfg.Logging.Level)
	defer logger.Sync()

	// Initialize driver service client
	driverServiceClient := service.NewDriverServiceClient(cfg.DriverService.BaseURL, logger)

	// Initialize handlers
	driverHandler := handler.NewDriverHandler(driverServiceClient, logger)
	authHandler := handler.NewAuthHandler(cfg, logger)

	// Initialize rate limiter
	rateLimiter := middleware.NewRateLimiter(&cfg.RateLimit, logger)

	// Setup router
	router := setupRouter(driverHandler, authHandler, cfg, logger, rateLimiter)

	// Start server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start server in a goroutine
	go func() {
		logger.Info("starting gateway", zap.String("port", cfg.Server.Port))
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

func setupRouter(
	driverHandler *handler.DriverHandler,
	authHandler *handler.AuthHandler,
	cfg *config.Config,
	logger *zap.Logger,
	rateLimiter *middleware.RateLimiter,
) *gin.Engine {
	if cfg.Logging.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Global middleware
	router.Use(middleware.CORS())
	router.Use(middleware.ErrorHandler(logger))
	router.Use(middleware.RequestLogger(logger))
	router.Use(rateLimiter.Limit())
	router.Use(gin.Recovery())

	// Swagger documentation (before other routes to avoid conflicts)
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	router.POST("/auth/login", authHandler.Login)

	// Driver routes
	drivers := router.Group("/drivers")
	{
		// Protected routes (require JWT)
		if cfg.JWT.Enabled {
			drivers.POST("", middleware.JWTAuth(cfg, logger), driverHandler.CreateDriver)
			drivers.PUT("/:id", middleware.JWTAuth(cfg, logger), driverHandler.UpdateDriver)
		} else {
			drivers.POST("", driverHandler.CreateDriver)
			drivers.PUT("/:id", driverHandler.UpdateDriver)
		}

		// Public routes (with optional API key protection)
		if cfg.APIKey.Enabled {
			// Apply API key to selected endpoints
			drivers.GET("/nearby", middleware.APIKeyAuth(cfg, logger), driverHandler.FindNearbyDrivers)
			drivers.GET("", middleware.APIKeyAuth(cfg, logger), driverHandler.ListDrivers)
			drivers.GET("/:id", driverHandler.GetDriver) // Keep this public
		} else {
			// All GET routes are public when API key is disabled
			drivers.GET("/:id", driverHandler.GetDriver)
			drivers.GET("", driverHandler.ListDrivers)
			drivers.GET("/nearby", driverHandler.FindNearbyDrivers)
		}
	}

	return router
}
