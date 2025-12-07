package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/bitaksi/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestNewAuthHandler(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret",
			Expiration: 24 * time.Hour,
		},
	}
	logger := zap.NewNop()
	handler := NewAuthHandler(cfg, logger)

	assert.NotNil(t, handler)
	assert.Equal(t, cfg, handler.config)
	assert.Equal(t, logger, handler.logger)
}

func TestAuthHandler_Login(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-for-testing",
			Expiration: 24 * time.Hour,
		},
	}
	logger := zap.NewNop()

	tests := []struct {
		name           string
		requestBody    interface{}
		expectedStatus int
		expectedError  string
		expectToken    bool
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "password",
			},
			expectedStatus: http.StatusOK,
			expectToken:    true,
		},
		{
			name: "empty username",
			requestBody: map[string]interface{}{
				"username": "",
				"password": "password",
			},
			expectedStatus: http.StatusBadRequest, // JSON binding validates first
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "empty password",
			requestBody: map[string]interface{}{
				"username": "admin",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest, // JSON binding validates first
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "both empty",
			requestBody: map[string]interface{}{
				"username": "",
				"password": "",
			},
			expectedStatus: http.StatusBadRequest, // JSON binding fails first
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name:           "invalid JSON",
			requestBody:    "invalid json",
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
		{
			name: "missing fields",
			requestBody: map[string]interface{}{
				"username": "admin",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "VALIDATION_ERROR",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := NewAuthHandler(cfg, logger)

			router := gin.New()
			gin.SetMode(gin.TestMode)
			router.POST("/auth/login", handler.Login)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest("POST", "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectToken {
				var response LoginResponse
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.NotEmpty(t, response.Token)
			}
			if tt.expectedError != "" {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.Equal(t, tt.expectedError, response["error"].(map[string]interface{})["code"])
			}
		})
	}
}

func TestAuthHandler_generateToken(t *testing.T) {
	cfg := &config.Config{
		JWT: config.JWTConfig{
			Secret:     "test-secret-key-for-testing",
			Expiration: 24 * time.Hour,
		},
	}
	logger := zap.NewNop()
	handler := NewAuthHandler(cfg, logger)

	token, err := handler.generateToken("testuser")
	assert.NoError(t, err)
	assert.NotEmpty(t, token)
}
