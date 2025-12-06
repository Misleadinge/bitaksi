package handler

import "github.com/gin-gonic/gin"

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

// respondError is a helper function to send error responses
func respondError(c *gin.Context, status int, code, message string) {
	var errResp ErrorResponse
	errResp.Error.Code = code
	errResp.Error.Message = message
	c.JSON(status, errResp)
}
