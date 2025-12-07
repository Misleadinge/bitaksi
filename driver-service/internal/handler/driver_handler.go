package handler

import (
	"net/http"
	"strconv"

	"github.com/bitaksi/driver-service/internal/domain"
	"github.com/bitaksi/driver-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DriverHandler handles HTTP requests for drivers
type DriverHandler struct {
	useCase usecase.DriverUseCase
	logger  *zap.Logger
}

// NewDriverHandler creates a new driver handler
func NewDriverHandler(useCase usecase.DriverUseCase, logger *zap.Logger) *DriverHandler {
	return &DriverHandler{
		useCase: useCase,
		logger:  logger,
	}
}

// CreateDriver handles POST /drivers
// @Summary Create a new driver
// @Description Create a new taxi driver
// @Tags drivers
// @Accept json
// @Produce json
// @Param driver body usecase.CreateDriverRequest true "Driver information" example({"firstName":"Ahmet","lastName":"Demir","plate":"34ABC123","taksiType":"sari","carBrand":"Toyota","carModel":"Corolla","lat":41.0431,"lon":29.0099})
// @Success 201 {object} domain.Driver "Driver created successfully" example({"id":"507f1f77bcf86cd799439011","firstName":"Ahmet","lastName":"Demir","plate":"34ABC123","taxiType":"sari","carBrand":"Toyota","carModel":"Corolla","location":{"lat":41.0431,"lon":29.0099},"createdAt":"2025-12-06T01:00:00Z","updatedAt":"2025-12-06T01:00:00Z"})
// @Failure 400 {object} ErrorResponse "Validation error" example({"error":{"code":"VALIDATION_ERROR","message":"plate must be in format: 2-3 digits, 1-3 letters, 1-4 digits (e.g., 34ABC123)"}})
// @Failure 500 {object} ErrorResponse "Internal server error" example({"error":{"code":"INTERNAL_ERROR","message":"failed to create driver"}})
// @Router /drivers [post]
func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var req usecase.CreateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	driver, err := h.useCase.CreateDriver(c.Request.Context(), &req)
	if err != nil {
		if isValidationError(err) {
			h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		h.logger.Error("failed to create driver", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create driver")
		return
	}

	c.JSON(http.StatusCreated, driver)
}

// UpdateDriver handles PUT /drivers/:id
// @Summary Update a driver
// @Description Update an existing driver. Location can be updated using top-level lat/lon fields (same format as create): {"lat": 41.0, "lon": 29.0}
// @Tags drivers
// @Accept json
// @Produce json
// @Param id path string true "Driver ID" example("507f1f77bcf86cd799439011")
// @Param driver body usecase.UpdateDriverRequest true "Driver update information. Location uses top-level lat/lon fields." example({"firstName":"Ali","lastName":"Kurt","plate":"34G99","taksiType":"siyah","carBrand":"Mercedes","carModel":"G Class","lat":42.0082,"lon":28.9784})
// @Success 200 {object} domain.Driver "Driver updated successfully" example({"id":"507f1f77bcf86cd799439011","firstName":"Ali","lastName":"Kurt","plate":"34G99","taxiType":"siyah","carBrand":"Mercedes","carModel":"G Class","location":{"lat":42.0082,"lon":28.9784},"createdAt":"2025-12-06T01:00:00Z","updatedAt":"2025-12-06T01:30:00Z"})
// @Failure 400 {object} ErrorResponse "Validation error" example({"error":{"code":"VALIDATION_ERROR","message":"both lat and lon must be provided together"}})
// @Failure 404 {object} ErrorResponse "Driver not found" example({"error":{"code":"NOT_FOUND","message":"driver not found"}})
// @Failure 500 {object} ErrorResponse "Internal server error" example({"error":{"code":"INTERNAL_ERROR","message":"failed to update driver"}})
// @Router /drivers/{id} [put]
func (h *DriverHandler) UpdateDriver(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "driver ID is required")
		return
	}

	var req usecase.UpdateDriverRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	driver, err := h.useCase.UpdateDriver(c.Request.Context(), id, &req)
	if err != nil {
		if err.Error() == "driver not found" {
			h.respondError(c, http.StatusNotFound, "NOT_FOUND", "driver not found")
			return
		}
		if isValidationError(err) {
			h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		h.logger.Error("failed to update driver", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update driver")
		return
	}

	c.JSON(http.StatusOK, driver)
}

// GetDriver handles GET /drivers/:id
// @Summary Get a driver by ID
// @Description Get driver details by ID
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID" example("507f1f77bcf86cd799439011")
// @Success 200 {object} domain.Driver "Driver details" example({"id":"507f1f77bcf86cd799439011","firstName":"Ahmet","lastName":"Demir","plate":"34ABC123","taxiType":"sari","carBrand":"Toyota","carModel":"Corolla","location":{"lat":41.0431,"lon":29.0099},"createdAt":"2025-12-06T01:00:00Z","updatedAt":"2025-12-06T01:00:00Z"})
// @Failure 404 {object} ErrorResponse "Driver not found" example({"error":{"code":"NOT_FOUND","message":"driver not found"}})
// @Failure 500 {object} ErrorResponse "Internal server error" example({"error":{"code":"INTERNAL_ERROR","message":"failed to get driver"}})
// @Router /drivers/{id} [get]
func (h *DriverHandler) GetDriver(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "driver ID is required")
		return
	}

	driver, err := h.useCase.GetDriver(c.Request.Context(), id)
	if err != nil {
		if err.Error() == "driver not found" {
			h.respondError(c, http.StatusNotFound, "NOT_FOUND", "driver not found")
			return
		}
		h.logger.Error("failed to get driver", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get driver")
		return
	}

	c.JSON(http.StatusOK, driver)
}

// ListDrivers handles GET /drivers
// @Summary List drivers
// @Description Get a paginated list of drivers
// @Tags drivers
// @Produce json
// @Param page query int false "Page number" default(1) example(1)
// @Param pageSize query int false "Page size" default(20) example(20)
// @Success 200 {object} usecase.ListDriversResponse "Paginated list of drivers" example({"drivers":[{"id":"507f1f77bcf86cd799439011","firstName":"Ahmet","lastName":"Demir","plate":"34ABC123","taxiType":"sari","carBrand":"Toyota","carModel":"Corolla","location":{"lat":41.0431,"lon":29.0099},"createdAt":"2025-12-06T01:00:00Z","updatedAt":"2025-12-06T01:00:00Z"}],"totalCount":1,"page":1,"pageSize":20})
// @Failure 400 {object} ErrorResponse "Validation error" example({"error":{"code":"VALIDATION_ERROR","message":"invalid page number"}})
// @Failure 500 {object} ErrorResponse "Internal server error" example({"error":{"code":"INTERNAL_ERROR","message":"failed to list drivers"}})
// @Router /drivers [get]
func (h *DriverHandler) ListDrivers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("pageSize", "20"))

	response, err := h.useCase.ListDrivers(c.Request.Context(), page, pageSize)
	if err != nil {
		h.logger.Error("failed to list drivers", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list drivers")
		return
	}

	c.JSON(http.StatusOK, response)
}

// FindNearbyDrivers handles GET /drivers/nearby
// @Summary Find nearby drivers
// @Description Find drivers within 6km radius
// @Tags drivers
// @Produce json
// @Param lat query float64 true "Latitude" example(41.0431)
// @Param lon query float64 true "Longitude" example(29.0099)
// @Param taksiType query string false "Taxi type (sari, turkuaz, siyah)" example(sari)
// @Success 200 {array} usecase.NearbyDriverResponse "List of nearby drivers sorted by distance" example([{"id":"507f1f77bcf86cd799439011","firstName":"Ahmet","lastName":"Demir","plate":"34ABC123","taxiType":"sari","carBrand":"Toyota","carModel":"Corolla","location":{"lat":41.0431,"lon":29.0099},"distance":0.5}])
// @Failure 400 {object} ErrorResponse "Validation error" example({"error":{"code":"VALIDATION_ERROR","message":"latitude is required"}})
// @Failure 500 {object} ErrorResponse "Internal server error" example({"error":{"code":"INTERNAL_ERROR","message":"failed to find nearby drivers"}})
// @Router /drivers/nearby [get]
func (h *DriverHandler) FindNearbyDrivers(c *gin.Context) {
	latStr := c.Query("lat")
	lonStr := c.Query("lon")
	taksiTypeStr := c.Query("taksiType")

	if latStr == "" || lonStr == "" {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "lat and lon are required")
		return
	}

	lat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid lat format")
		return
	}

	lon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid lon format")
		return
	}

	var taxiType *domain.TaxiType
	if taksiTypeStr != "" {
		tt := domain.TaxiType(taksiTypeStr)
		if !tt.IsValid() {
			h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "invalid taksiType. Must be one of: sari, turkuaz, siyah")
			return
		}
		taxiType = &tt
	}

	drivers, err := h.useCase.FindNearbyDrivers(c.Request.Context(), lat, lon, taxiType)
	if err != nil {
		if isValidationError(err) {
			h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
			return
		}
		h.logger.Error("failed to find nearby drivers", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to find nearby drivers")
		return
	}

	c.JSON(http.StatusOK, drivers)
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code" example:"VALIDATION_ERROR"`
		Message string `json:"message" example:"plate must be in format: 2-3 digits, 1-3 letters, 1-4 digits (e.g., 34ABC123)"`
	} `json:"error"`
}

func (h *DriverHandler) respondError(c *gin.Context, status int, code, message string) {
	var errResp ErrorResponse
	errResp.Error.Code = code
	errResp.Error.Message = message
	c.JSON(status, errResp)
}

func isValidationError(err error) bool {
	return err != nil && (err.Error() == "firstName is required" ||
		err.Error() == "lastName is required" ||
		err.Error() == "plate is required" ||
		err.Error() == "carBrand is required" ||
		err.Error() == "carModel is required" ||
		err.Error() == "latitude must be between -90 and 90" ||
		err.Error() == "longitude must be between -180 and 180" ||
		err.Error() == "driver not found" ||
		err.Error() == "invalid driver ID")
}
