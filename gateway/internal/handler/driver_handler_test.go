package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bitaksi/gateway/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// mockDriverServiceClient removed - using httptest.Server instead

func setupGatewayRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	return router
}

func createMockResponse(statusCode int, body string) *http.Response {
	return &http.Response{
		StatusCode: statusCode,
		Body:       io.NopCloser(bytes.NewBufferString(body)),
		Header:     make(http.Header),
	}
}

func TestNewDriverHandler(t *testing.T) {
	logger := zap.NewNop()
	realService := service.NewDriverServiceClient("http://localhost:8081", logger)
	handler := NewDriverHandler(realService, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, realService, handler.driverService)
	assert.Equal(t, logger, handler.logger)
}

func TestDriverHandler_CreateDriver(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		requestBody    interface{}
		mockFunc       func(body interface{}) (*http.Response, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful create",
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
			mockFunc: func(body interface{}) (*http.Response, error) {
				return createMockResponse(http.StatusCreated, `{"id":"test-id","firstName":"Ahmet"}`), nil
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "service error",
			requestBody: map[string]interface{}{
				"firstName": "Ahmet",
			},
			mockFunc:       nil, // No server = connection error
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Use httptest server to mock the driver service
			var mockServer *httptest.Server
			if tt.mockFunc != nil {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					resp, _ := tt.mockFunc(nil)
					if resp != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(resp.StatusCode)
						io.Copy(w, resp.Body)
					}
				}))
				defer mockServer.Close()
			} else if tt.name == "service error" {
				// For service error, use invalid URL to simulate connection failure
				mockServer = nil
			} else {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))
				defer mockServer.Close()
			}

			baseURL := "http://invalid-host:9999" // Will fail to connect
			if mockServer != nil {
				baseURL = mockServer.URL
			}
			realService := service.NewDriverServiceClient(baseURL, logger)
			handler := NewDriverHandler(realService, logger)

			router := setupGatewayRouter()
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
		mockFunc       func(id string, body interface{}) (*http.Response, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful update",
			id:   "test-id",
			requestBody: map[string]interface{}{
				"firstName": "Mehmet",
			},
			mockFunc: func(id string, body interface{}) (*http.Response, error) {
				return createMockResponse(http.StatusOK, `{"id":"test-id","firstName":"Mehmet"}`), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid JSON",
			id:             "test-id",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "service error",
			id:   "test-id",
			requestBody: map[string]interface{}{
				"firstName": "Mehmet",
			},
			mockFunc:       nil, // No server = connection error
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockServer *httptest.Server
			baseURL := "http://invalid-host:9999"
			if tt.mockFunc != nil {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					// Extract ID from path
					id := r.URL.Path[len("/api/v1/drivers/"):]
					resp, _ := tt.mockFunc(id, nil)
					if resp != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(resp.StatusCode)
						io.Copy(w, resp.Body)
					}
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			} else if tt.name == "service error" {
				// For service error, don't create a server - use invalid URL to force connection error
				baseURL = "http://invalid-host-that-does-not-exist:9999"
			} else {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			}

			realService := service.NewDriverServiceClient(baseURL, logger)
			handler := NewDriverHandler(realService, logger)

			router := setupGatewayRouter()
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
		mockFunc       func(id string) (*http.Response, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful get",
			id:   "test-id",
			mockFunc: func(id string) (*http.Response, error) {
				return createMockResponse(http.StatusOK, `{"id":"test-id","firstName":"Ahmet"}`), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			id:             "test-id",
			mockFunc:       nil, // No server = connection error
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockServer *httptest.Server
			baseURL := "http://invalid-host:9999"
			if tt.mockFunc != nil {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					id := r.URL.Path[len("/api/v1/drivers/"):]
					resp, _ := tt.mockFunc(id)
					if resp != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(resp.StatusCode)
						io.Copy(w, resp.Body)
					}
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			} else if tt.name == "service error" {
				// For service error, don't create a server - use invalid URL to force connection error
				baseURL = "http://invalid-host-that-does-not-exist:9999"
			} else {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			}

			realService := service.NewDriverServiceClient(baseURL, logger)
			handler := NewDriverHandler(realService, logger)

			router := setupGatewayRouter()
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
		mockFunc       func(page, pageSize string) (*http.Response, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful list",
			queryParams: "?page=1&pageSize=20",
			mockFunc: func(page, pageSize string) (*http.Response, error) {
				return createMockResponse(http.StatusOK, `{"drivers":[],"totalCount":0}`), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful list with empty params",
			queryParams: "",
			mockFunc: func(page, pageSize string) (*http.Response, error) {
				return createMockResponse(http.StatusOK, `{"drivers":[],"totalCount":0}`), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "service error",
			queryParams:    "?page=1&pageSize=20",
			mockFunc:       nil, // No server = connection error
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockServer *httptest.Server
			baseURL := "http://invalid-host:9999"
			if tt.mockFunc != nil {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					page := r.URL.Query().Get("page")
					pageSize := r.URL.Query().Get("pageSize")
					resp, _ := tt.mockFunc(page, pageSize)
					if resp != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(resp.StatusCode)
						io.Copy(w, resp.Body)
					}
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			} else if tt.name == "service error" {
				// For service error, don't create a server - use invalid URL to force connection error
				baseURL = "http://invalid-host-that-does-not-exist:9999"
			} else {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			}

			realService := service.NewDriverServiceClient(baseURL, logger)
			handler := NewDriverHandler(realService, logger)

			router := setupGatewayRouter()
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
		mockFunc       func(lat, lon, taksiType string) (*http.Response, error)
		expectedStatus int
		expectedError  string
	}{
		{
			name:        "successful find nearby",
			queryParams: "?lat=41.0431&lon=29.0099",
			mockFunc: func(lat, lon, taksiType string) (*http.Response, error) {
				return createMockResponse(http.StatusOK, `[]`), nil
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:        "successful find nearby with taksiType",
			queryParams: "?lat=41.0431&lon=29.0099&taksiType=sari",
			mockFunc: func(lat, lon, taksiType string) (*http.Response, error) {
				return createMockResponse(http.StatusOK, `[]`), nil
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
			name:           "service error",
			queryParams:    "?lat=41.0431&lon=29.0099",
			mockFunc:       nil, // No server = connection error
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var mockServer *httptest.Server
			baseURL := "http://invalid-host:9999"
			if tt.mockFunc != nil {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					lat := r.URL.Query().Get("lat")
					lon := r.URL.Query().Get("lon")
					taksiType := r.URL.Query().Get("taksiType")
					resp, _ := tt.mockFunc(lat, lon, taksiType)
					if resp != nil {
						w.Header().Set("Content-Type", "application/json")
						w.WriteHeader(resp.StatusCode)
						io.Copy(w, resp.Body)
					}
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			} else if tt.name == "service error" {
				// For service error, don't create a server - use invalid URL to force connection error
				baseURL = "http://invalid-host-that-does-not-exist:9999"
			} else if tt.name != "missing lat" && tt.name != "missing lon" {
				mockServer = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusBadRequest)
				}))
				defer mockServer.Close()
				baseURL = mockServer.URL
			}

			realService := service.NewDriverServiceClient(baseURL, logger)
			handler := NewDriverHandler(realService, logger)

			router := setupGatewayRouter()
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

func TestDriverHandler_forwardResponse(t *testing.T) {
	logger := zap.NewNop()
	realService := service.NewDriverServiceClient("http://localhost:8081", logger)
	handler := NewDriverHandler(realService, logger)

	tests := []struct {
		name           string
		response       *http.Response
		expectedStatus int
		expectedError  string
	}{
		{
			name: "successful forward",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"test":"data"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "error reading body",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       &errorReader{},
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
			expectedStatus: http.StatusInternalServerError,
			expectedError:  "INTERNAL_ERROR",
		},
		{
			name: "forward with multiple headers",
			response: &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(bytes.NewBufferString(`{"test":"data"}`)),
				Header: http.Header{
					"Content-Type": []string{"application/json"},
					"X-Custom-1":   []string{"value1", "value2"},
					"X-Custom-2":   []string{"value3"},
				},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "forward with different status code",
			response: &http.Response{
				StatusCode: http.StatusCreated,
				Body:       io.NopCloser(bytes.NewBufferString(`{"id":"123"}`)),
				Header:     http.Header{"Content-Type": []string{"application/json"}},
			},
			expectedStatus: http.StatusCreated,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router := setupGatewayRouter()
			router.GET("/test", func(c *gin.Context) {
				handler.forwardResponse(c, tt.response)
			})

			req := httptest.NewRequest("GET", "/test", nil)
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

// errorReader is a reader that always returns an error
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func (e *errorReader) Close() error {
	return nil
}

func TestDriverHandler_respondError(t *testing.T) {
	logger := zap.NewNop()
	realService := service.NewDriverServiceClient("http://localhost:8081", logger)
	handler := NewDriverHandler(realService, logger)

	router := setupGatewayRouter()
	router.GET("/test", func(c *gin.Context) {
		handler.respondError(c, http.StatusBadRequest, "TEST_ERROR", "test message")
	})

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)
	errorObj := response["error"].(map[string]interface{})
	assert.Equal(t, "TEST_ERROR", errorObj["code"])
	assert.Equal(t, "test message", errorObj["message"])
}
