package handler

import (
	"io"
	"net/http"

	"github.com/bitaksi/gateway/internal/service"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// DriverHandler handles HTTP requests for drivers in the gateway
type DriverHandler struct {
	driverService *service.DriverServiceClient
	logger        *zap.Logger
}

// NewDriverHandler creates a new driver handler
func NewDriverHandler(driverService *service.DriverServiceClient, logger *zap.Logger) *DriverHandler {
	return &DriverHandler{
		driverService: driverService,
		logger:        logger,
	}
}

// CreateDriver handles POST /drivers
// @Summary Create a new driver
// @Description Create a new taxi driver
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param driver body CreateDriverRequest true "Driver information"
// @Success 201 {object} Driver "Driver created successfully"
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /drivers [post]
func (h *DriverHandler) CreateDriver(c *gin.Context) {
	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.driverService.CreateDriver(body)
	if err != nil {
		h.logger.Error("failed to forward create driver request", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to create driver")
		return
	}
	defer resp.Body.Close()

	h.forwardResponse(c, resp)
}

// UpdateDriver handles PUT /drivers/:id
// @Summary Update a driver
// @Description Update an existing driver
// @Tags drivers
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Driver ID"
// @Param driver body UpdateDriverRequest true "Driver update information"
// @Success 200 {object} Driver "Driver updated successfully"
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 404 {object} ErrorResponse "Driver not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /drivers/{id} [put]
func (h *DriverHandler) UpdateDriver(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "driver ID is required")
		return
	}

	var body map[string]interface{}
	if err := c.ShouldBindJSON(&body); err != nil {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	resp, err := h.driverService.UpdateDriver(id, body)
	if err != nil {
		h.logger.Error("failed to forward update driver request", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to update driver")
		return
	}
	defer resp.Body.Close()

	h.forwardResponse(c, resp)
}

// GetDriver handles GET /drivers/:id
// @Summary Get a driver by ID
// @Description Get driver details by ID
// @Tags drivers
// @Produce json
// @Param id path string true "Driver ID"
// @Success 200 {object} Driver "Driver details"
// @Failure 404 {object} ErrorResponse "Driver not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /drivers/{id} [get]
func (h *DriverHandler) GetDriver(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "driver ID is required")
		return
	}

	resp, err := h.driverService.GetDriver(id)
	if err != nil {
		h.logger.Error("failed to forward get driver request", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to get driver")
		return
	}
	defer resp.Body.Close()

	h.forwardResponse(c, resp)
}

// ListDrivers handles GET /drivers
// @Summary List drivers
// @Description Get a paginated list of drivers
// @Tags drivers
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param pageSize query int false "Page size" default(20)
// @Success 200 {object} ListDriversResponse "Paginated list of drivers"
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /drivers [get]
func (h *DriverHandler) ListDrivers(c *gin.Context) {
	page := c.DefaultQuery("page", "")
	pageSize := c.DefaultQuery("pageSize", "")

	resp, err := h.driverService.ListDrivers(page, pageSize)
	if err != nil {
		h.logger.Error("failed to forward list drivers request", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to list drivers")
		return
	}
	defer resp.Body.Close()

	h.forwardResponse(c, resp)
}

// FindNearbyDrivers handles GET /drivers/nearby
// @Summary Find nearby drivers
// @Description Find drivers within 6km radius
// @Tags drivers
// @Produce json
// @Param lat query float64 true "Latitude"
// @Param lon query float64 true "Longitude"
// @Param taksiType query string false "Taxi type (sari, turkuaz, siyah)"
// @Success 200 {array} NearbyDriverResponse "List of nearby drivers sorted by distance"
// @Failure 400 {object} ErrorResponse "Validation error"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /drivers/nearby [get]
func (h *DriverHandler) FindNearbyDrivers(c *gin.Context) {
	lat := c.Query("lat")
	lon := c.Query("lon")
	taksiType := c.Query("taksiType")

	if lat == "" || lon == "" {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", "lat and lon are required")
		return
	}

	resp, err := h.driverService.FindNearbyDrivers(lat, lon, taksiType)
	if err != nil {
		h.logger.Error("failed to forward find nearby drivers request", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to find nearby drivers")
		return
	}
	defer resp.Body.Close()

	h.forwardResponse(c, resp)
}

// forwardResponse forwards the response from the driver service to the client
func (h *DriverHandler) forwardResponse(c *gin.Context, resp *http.Response) {
	// Copy status code
	c.Status(resp.StatusCode)

	// Copy headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Copy body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		h.logger.Error("failed to read response body", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to read response")
		return
	}

	c.Data(resp.StatusCode, resp.Header.Get("Content-Type"), body)
}

func (h *DriverHandler) respondError(c *gin.Context, status int, code, message string) {
	respondError(c, status, code, message)
}
