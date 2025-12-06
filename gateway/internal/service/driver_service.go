package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// DriverServiceClient handles communication with the driver service
type DriverServiceClient struct {
	baseURL    string
	httpClient *http.Client
	logger     *zap.Logger
}

// NewDriverServiceClient creates a new driver service client
func NewDriverServiceClient(baseURL string, logger *zap.Logger) *DriverServiceClient {
	return &DriverServiceClient{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		logger: logger,
	}
}

// CreateDriver forwards a create driver request to the driver service
func (c *DriverServiceClient) CreateDriver(body interface{}) (*http.Response, error) {
	return c.doRequest("POST", "/api/v1/drivers", body)
}

// UpdateDriver forwards an update driver request to the driver service
func (c *DriverServiceClient) UpdateDriver(id string, body interface{}) (*http.Response, error) {
	return c.doRequest("PUT", fmt.Sprintf("/api/v1/drivers/%s", id), body)
}

// GetDriver forwards a get driver request to the driver service
func (c *DriverServiceClient) GetDriver(id string) (*http.Response, error) {
	return c.doRequest("GET", fmt.Sprintf("/api/v1/drivers/%s", id), nil)
}

// ListDrivers forwards a list drivers request to the driver service
func (c *DriverServiceClient) ListDrivers(page, pageSize string) (*http.Response, error) {
	url := "/api/v1/drivers"
	if page != "" || pageSize != "" {
		url += "?"
		if page != "" {
			url += "page=" + page
		}
		if pageSize != "" {
			if page != "" {
				url += "&"
			}
			url += "pageSize=" + pageSize
		}
	}
	return c.doRequest("GET", url, nil)
}

// FindNearbyDrivers forwards a find nearby drivers request to the driver service
func (c *DriverServiceClient) FindNearbyDrivers(lat, lon, taksiType string) (*http.Response, error) {
	url := fmt.Sprintf("/api/v1/drivers/nearby?lat=%s&lon=%s", lat, lon)
	if taksiType != "" {
		url += "&taksiType=" + taksiType
	}
	return c.doRequest("GET", url, nil)
}

func (c *DriverServiceClient) doRequest(method, path string, body interface{}) (*http.Response, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	c.logger.Debug("forwarding request to driver service",
		zap.String("method", method),
		zap.String("url", url),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("failed to forward request to driver service",
			zap.Error(err),
			zap.String("method", method),
			zap.String("url", url),
		)
		return nil, fmt.Errorf("failed to forward request: %w", err)
	}

	return resp, nil
}
