package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitaksi/driver-service/internal/domain"
	"github.com/bitaksi/driver-service/internal/usecase"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockDriverUseCase is a mock implementation of DriverUseCase
type mockDriverUseCase struct {
	createDriverFunc      func(ctx context.Context, req *usecase.CreateDriverRequest) (*domain.Driver, error)
	updateDriverFunc      func(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error)
	getDriverFunc         func(ctx context.Context, id string) (*domain.Driver, error)
	listDriversFunc       func(ctx context.Context, page, pageSize int) (*usecase.ListDriversResponse, error)
	findNearbyDriversFunc func(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*usecase.NearbyDriverResponse, error)
}

func (m *mockDriverUseCase) CreateDriver(ctx context.Context, req *usecase.CreateDriverRequest) (*domain.Driver, error) {
	if m.createDriverFunc != nil {
		return m.createDriverFunc(ctx, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDriverUseCase) UpdateDriver(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error) {
	if m.updateDriverFunc != nil {
		return m.updateDriverFunc(ctx, id, req)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDriverUseCase) GetDriver(ctx context.Context, id string) (*domain.Driver, error) {
	if m.getDriverFunc != nil {
		return m.getDriverFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDriverUseCase) ListDrivers(ctx context.Context, page, pageSize int) (*usecase.ListDriversResponse, error) {
	if m.listDriversFunc != nil {
		return m.listDriversFunc(ctx, page, pageSize)
	}
	return nil, errors.New("not implemented")
}

func (m *mockDriverUseCase) FindNearbyDrivers(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*usecase.NearbyDriverResponse, error) {
	if m.findNearbyDriversFunc != nil {
		return m.findNearbyDriversFunc(ctx, lat, lon, taxiType)
	}
	return nil, errors.New("not implemented")
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestNewDriverHandler(t *testing.T) {
	logger := zap.NewNop()
	mockUC := &mockDriverUseCase{}
	handler := NewDriverHandler(mockUC, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, mockUC, handler.useCase)
	assert.Equal(t, logger, handler.logger)
}

func TestDriverHandler_CreateDriver(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, req *usecase.CreateDriverRequest) (*domain.Driver, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful creation",
			requestBody: map[string]interface{}{
				"firstName": "Ahmet",
				"lastName":  "Demir",
				"plate":     "34ABC123",
				"taksiType": "sari",
				"carBrand":  "Toyota",
				"carModel":  "Corolla",
				"lat":       41.0431,
				"lon":       29.0099,
			},
			mockFunc: func(ctx context.Context, req *usecase.CreateDriverRequest) (*domain.Driver, error) {
				return &domain.Driver{
					ID:        "test-id",
					FirstName: req.FirstName,
					LastName:  req.LastName,
					Plate:     req.Plate,
					TaxiType:  req.TaxiType,
					CarBrand:  req.CarBrand,
					CarModel:  req.CarModel,
					Location: domain.Location{
						Lat: req.Lat,
						Lon: req.Lon,
					},
				}, nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "validation error",
			requestBody: map[string]interface{}{
				"firstName": "",
			},
			mockFunc: func(ctx context.Context, req *usecase.CreateDriverRequest) (*domain.Driver, error) {
				return nil, errors.New("firstName is required")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "internal error",
			requestBody: map[string]interface{}{
				"firstName": "Ahmet",
				"lastName":  "Demir",
				"plate":     "34ABC123",
				"taksiType": "sari",
				"carBrand":  "Toyota",
				"carModel":  "Corolla",
				"lat":       41.0431,
				"lon":       29.0099,
			},
			mockFunc: func(ctx context.Context, req *usecase.CreateDriverRequest) (*domain.Driver, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockDriverUseCase{
				createDriverFunc: tt.mockFunc,
			}
			handler := NewDriverHandler(mockUC, logger)

			router := setupRouter()
			router.POST("/drivers", handler.CreateDriver)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/drivers", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" && w.Body.Len() > 0 {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if errorObj, ok := response["error"].(map[string]interface{}); ok {
						assert.Equal(t, tt.expectedError, errorObj["code"])
					}
				}
			}
		})
	}
}

func TestDriverHandler_UpdateDriver(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		id             string
		requestBody    interface{}
		mockFunc       func(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful update",
			id:   "test-id",
			requestBody: map[string]interface{}{
				"firstName": "Mehmet",
			},
			mockFunc: func(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error) {
				return &domain.Driver{
					ID:        id,
					FirstName: "Mehmet",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "missing id",
			id:          "",
			requestBody: map[string]interface{}{"firstName": "Mehmet"},
			mockFunc: func(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error) {
				return nil, nil
			}, // Should not be called
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "driver not found",
			id:   "non-existent",
			requestBody: map[string]interface{}{
				"firstName": "Mehmet",
			},
			mockFunc: func(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error) {
				return nil, errors.New("driver not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name: "validation error",
			id:   "test-id",
			requestBody: map[string]interface{}{
				"plate": "INVALID",
			},
			mockFunc: func(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error) {
				return nil, errors.New("plate must be in format")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "internal error",
			id:   "test-id",
			requestBody: map[string]interface{}{
				"firstName": "Mehmet",
			},
			mockFunc: func(ctx context.Context, id string, req *usecase.UpdateDriverRequest) (*domain.Driver, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
		{
			name:           "invalid JSON",
			id:             "test-id",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockDriverUseCase{}
			if tt.mockFunc != nil {
				mockUC.updateDriverFunc = tt.mockFunc
			}
			handler := NewDriverHandler(mockUC, logger)

			router := setupRouter()
			router.PUT("/drivers/:id", handler.UpdateDriver)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("PUT", "/drivers/"+tt.id, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" && w.Body.Len() > 0 {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if errorObj, ok := response["error"].(map[string]interface{}); ok {
						assert.Equal(t, tt.expectedError, errorObj["code"])
					}
				}
			}
		})
	}
}

func TestDriverHandler_GetDriver(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		id             string
		mockFunc       func(ctx context.Context, id string) (*domain.Driver, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful get",
			id:   "test-id",
			mockFunc: func(ctx context.Context, id string) (*domain.Driver, error) {
				return &domain.Driver{
					ID:        id,
					FirstName: "Ahmet",
					LastName:  "Demir",
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing id",
			id:             "",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "driver not found",
			id:   "non-existent",
			mockFunc: func(ctx context.Context, id string) (*domain.Driver, error) {
				return nil, errors.New("driver not found")
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "NOT_FOUND",
		},
		{
			name: "internal error",
			id:   "test-id",
			mockFunc: func(ctx context.Context, id string) (*domain.Driver, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockDriverUseCase{
				getDriverFunc: tt.mockFunc,
			}
			handler := NewDriverHandler(mockUC, logger)

			router := setupRouter()
			router.GET("/drivers/:id", handler.GetDriver)

			req := httptest.NewRequest("GET", "/drivers/"+tt.id, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" && w.Body.Len() > 0 {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if errorObj, ok := response["error"].(map[string]interface{}); ok {
						assert.Equal(t, tt.expectedError, errorObj["code"])
					}
				}
			}
		})
	}
}

func TestDriverHandler_ListDrivers(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		queryParams    string
		mockFunc       func(ctx context.Context, page, pageSize int) (*usecase.ListDriversResponse, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful list",
			queryParams: "?page=1&pageSize=20",
			mockFunc: func(ctx context.Context, page, pageSize int) (*usecase.ListDriversResponse, error) {
				return &usecase.ListDriversResponse{
					Drivers:    []*domain.Driver{},
					TotalCount: 0,
					Page:       page,
					PageSize:   pageSize,
				}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "invalid page",
			queryParams: "?page=invalid",
			mockFunc: func(ctx context.Context, page, pageSize int) (*usecase.ListDriversResponse, error) {
				return &usecase.ListDriversResponse{}, nil
			},
			expectedStatus: http.StatusOK, // Page defaults to 1
		},
		{
			name:        "invalid pageSize",
			queryParams: "?pageSize=invalid",
			mockFunc: func(ctx context.Context, page, pageSize int) (*usecase.ListDriversResponse, error) {
				return &usecase.ListDriversResponse{}, nil
			},
			expectedStatus: http.StatusOK, // PageSize defaults to 20
		},
		{
			name:        "internal error",
			queryParams: "?page=1&pageSize=20",
			mockFunc: func(ctx context.Context, page, pageSize int) (*usecase.ListDriversResponse, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockDriverUseCase{
				listDriversFunc: tt.mockFunc,
			}
			handler := NewDriverHandler(mockUC, logger)

			router := setupRouter()
			router.GET("/drivers", handler.ListDrivers)

			req := httptest.NewRequest("GET", "/drivers"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" && w.Body.Len() > 0 {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if errorObj, ok := response["error"].(map[string]interface{}); ok {
						assert.Equal(t, tt.expectedError, errorObj["code"])
					}
				}
			}
		})
	}
}

func TestDriverHandler_FindNearbyDrivers(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		queryParams    string
		mockFunc       func(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*usecase.NearbyDriverResponse, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful find nearby",
			queryParams: "?lat=41.0431&lon=29.0099",
			mockFunc: func(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*usecase.NearbyDriverResponse, error) {
				return []*usecase.NearbyDriverResponse{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "with taxi type filter",
			queryParams: "?lat=41.0431&lon=29.0099&taksiType=sari",
			mockFunc: func(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*usecase.NearbyDriverResponse, error) {
				return []*usecase.NearbyDriverResponse{}, nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "missing lat",
			queryParams:    "?lon=29.0099",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "missing lon",
			queryParams:    "?lat=41.0431",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "invalid lat format",
			queryParams:    "?lat=invalid&lon=29.0099",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "invalid lon format",
			queryParams:    "?lat=41.0431&lon=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "invalid taxi type",
			queryParams:    "?lat=41.0431&lon=29.0099&taksiType=invalid",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:        "validation error from use case",
			queryParams: "?lat=41.0431&lon=29.0099",
			mockFunc: func(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*usecase.NearbyDriverResponse, error) {
				return nil, errors.New("latitude must be between -90 and 90")
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:        "internal error",
			queryParams: "?lat=41.0431&lon=29.0099",
			mockFunc: func(ctx context.Context, lat, lon float64, taxiType *domain.TaxiType) ([]*usecase.NearbyDriverResponse, error) {
				return nil, errors.New("database error")
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockUC := &mockDriverUseCase{
				findNearbyDriversFunc: tt.mockFunc,
			}
			handler := NewDriverHandler(mockUC, logger)

			router := setupRouter()
			router.GET("/drivers/nearby", handler.FindNearbyDrivers)

			req := httptest.NewRequest("GET", "/drivers/nearby"+tt.queryParams, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedError != "" && w.Body.Len() > 0 {
				var response map[string]interface{}
				if err := json.Unmarshal(w.Body.Bytes(), &response); err == nil {
					if errorObj, ok := response["error"].(map[string]interface{}); ok {
						assert.Equal(t, tt.expectedError, errorObj["code"])
					}
				}
			}
		})
	}
}

func TestDriverHandler_respondError(t *testing.T) {
	logger := zap.NewNop()
	handler := NewDriverHandler(&mockDriverUseCase{}, logger)

	router := setupRouter()
	router.GET("/test", func(c *gin.Context) {
		handler.respondError(c, http.StatusBadRequest, "TEST_ERROR", "test message")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	assert.Equal(t, "TEST_ERROR", response["error"].(map[string]interface{})["code"])
	assert.Equal(t, "test message", response["error"].(map[string]interface{})["message"])
}

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "validation error - firstName",
			err:      errors.New("firstName is required"),
			expected: true,
		},
		{
			name:     "validation error - lastName",
			err:      errors.New("lastName is required"),
			expected: true,
		},
		{
			name:     "validation error - plate",
			err:      errors.New("plate is required"),
			expected: true,
		},
		{
			name:     "validation error - carBrand",
			err:      errors.New("carBrand is required"),
			expected: true,
		},
		{
			name:     "validation error - carModel",
			err:      errors.New("carModel is required"),
			expected: true,
		},
		{
			name:     "validation error - latitude",
			err:      errors.New("latitude must be between -90 and 90"),
			expected: true,
		},
		{
			name:     "validation error - longitude",
			err:      errors.New("longitude must be between -180 and 180"),
			expected: true,
		},
		{
			name:     "not validation error",
			err:      errors.New("database error"),
			expected: false,
		},
		{
			name:     "driver not found",
			err:      errors.New("driver not found"),
			expected: true, // This is also considered validation error in the function
		},
		{
			name:     "invalid driver ID",
			err:      errors.New("invalid driver ID"),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidationError(tt.err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
