package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewDriverServiceClient(t *testing.T) {
	logger := zap.NewNop()
	client := NewDriverServiceClient("http://localhost:8081", logger)

	assert.NotNil(t, client)
	assert.Equal(t, "http://localhost:8081", client.baseURL)
	assert.NotNil(t, client.httpClient)
	assert.Equal(t, logger, client.logger)
}

func TestDriverServiceClient_CreateDriver(t *testing.T) {
	logger := zap.NewNop()

	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "/api/v1/drivers", r.URL.Path)

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "test-id",
			"firstName": "Ahmet",
		})
	}))
	defer server.Close()

	client := NewDriverServiceClient(server.URL, logger)
	body := map[string]interface{}{
		"firstName": "Ahmet",
		"lastName":  "Demir",
	}

	resp, err := client.CreateDriver(body)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	defer resp.Body.Close()
}

func TestDriverServiceClient_UpdateDriver(t *testing.T) {
	logger := zap.NewNop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PUT", r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/drivers/")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "test-id",
			"firstName": "Mehmet",
		})
	}))
	defer server.Close()

	client := NewDriverServiceClient(server.URL, logger)
	body := map[string]interface{}{
		"firstName": "Mehmet",
	}

	resp, err := client.UpdateDriver("test-id", body)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
}

func TestDriverServiceClient_GetDriver(t *testing.T) {
	logger := zap.NewNop()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "GET", r.Method)
		assert.Contains(t, r.URL.Path, "/api/v1/drivers/")

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":        "test-id",
			"firstName": "Ahmet",
		})
	}))
	defer server.Close()

	client := NewDriverServiceClient(server.URL, logger)
	resp, err := client.GetDriver("test-id")
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	defer resp.Body.Close()
}

func TestDriverServiceClient_ListDrivers(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name     string
		page     string
		pageSize string
		expected string
	}{
		{
			name:     "with pagination",
			page:     "1",
			pageSize: "20",
			expected: "/api/v1/drivers?page=1&pageSize=20",
		},
		{
			name:     "only page",
			page:     "1",
			pageSize: "",
			expected: "/api/v1/drivers?page=1",
		},
		{
			name:     "only pageSize",
			page:     "",
			pageSize: "20",
			expected: "/api/v1/drivers?pageSize=20",
		},
		{
			name:     "no pagination",
			page:     "",
			pageSize: "",
			expected: "/api/v1/drivers",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Equal(t, tt.expected, r.URL.Path+"?"+r.URL.RawQuery)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"drivers":    []interface{}{},
					"totalCount": 0,
				})
			}))
			defer server.Close()

			client := NewDriverServiceClient(server.URL, logger)
			resp, err := client.ListDrivers(tt.page, tt.pageSize)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			defer resp.Body.Close()
		})
	}
}

func TestDriverServiceClient_FindNearbyDrivers(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name      string
		lat       string
		lon       string
		taksiType string
		expected  string
	}{
		{
			name:      "with taxi type",
			lat:       "41.0431",
			lon:       "29.0099",
			taksiType: "sari",
			expected:  "/api/v1/drivers/nearby?lat=41.0431&lon=29.0099&taksiType=sari",
		},
		{
			name:      "without taxi type",
			lat:       "41.0431",
			lon:       "29.0099",
			taksiType: "",
			expected:  "/api/v1/drivers/nearby?lat=41.0431&lon=29.0099",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "GET", r.Method)
				assert.Contains(t, r.URL.String(), tt.expected)

				w.WriteHeader(http.StatusOK)
				json.NewEncoder(w).Encode([]interface{}{})
			}))
			defer server.Close()

			client := NewDriverServiceClient(server.URL, logger)
			resp, err := client.FindNearbyDrivers(tt.lat, tt.lon, tt.taksiType)
			assert.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, http.StatusOK, resp.StatusCode)
			defer resp.Body.Close()
		})
	}
}

func TestDriverServiceClient_doRequest(t *testing.T) {
	logger := zap.NewNop()

	tests := []struct {
		name           string
		method         string
		path           string
		body           interface{}
		handler        http.HandlerFunc
		expectedStatus int
		wantErr        bool
	}{
		{
			name:   "successful POST",
			method: "POST",
			path:   "/test",
			body:   map[string]string{"test": "data"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, "POST", r.Method)
				w.WriteHeader(http.StatusOK)
			},
			expectedStatus: http.StatusOK,
			wantErr:        false,
		},
		{
			name:   "server error",
			method: "GET",
			path:   "/test",
			body:   nil,
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedStatus: http.StatusInternalServerError,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			client := NewDriverServiceClient(server.URL, logger)
			resp, err := client.doRequest(tt.method, tt.path, tt.body)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedStatus, resp.StatusCode)
				if resp != nil {
					defer resp.Body.Close()
				}
			}
		})
	}
}
