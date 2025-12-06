package handler

import (
	"net/http"
	"time"

	"github.com/bitaksi/gateway/internal/config"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

// AuthHandler handles authentication requests
type AuthHandler struct {
	config *config.Config
	logger *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(cfg *config.Config, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		config: cfg,
		logger: logger,
	}
}

// LoginRequest represents a login request
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse represents a login response
type LoginResponse struct {
	Token string `json:"token"`
}

// Login handles POST /auth/login
// @Summary Login
// @Description Authenticate and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param credentials body LoginRequest true "Login credentials"
// @Success 200 {object} LoginResponse "Authentication successful"
// @Failure 401 {object} ErrorResponse "Unauthorized - invalid credentials"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "VALIDATION_ERROR", err.Error())
		return
	}

	// Simple mock authentication (in production, use proper user database)
	// For demo purposes, accept any username/password or use hardcoded admin
	if req.Username == "" || req.Password == "" {
		h.respondError(c, http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials")
		return
	}

	// Generate JWT token
	token, err := h.generateToken(req.Username)
	if err != nil {
		h.logger.Error("failed to generate token", zap.Error(err))
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", "failed to generate token")
		return
	}

	c.JSON(http.StatusOK, LoginResponse{Token: token})
}

// generateToken generates a JWT token for the user
func (h *AuthHandler) generateToken(username string) (string, error) {
	claims := jwt.MapClaims{
		"username": username,
		"exp":      time.Now().Add(h.config.JWT.Expiration).Unix(),
		"iat":      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.config.JWT.Secret))
}

func (h *AuthHandler) respondError(c *gin.Context, status int, code, message string) {
	respondError(c, status, code, message)
}
