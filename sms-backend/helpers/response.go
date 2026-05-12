package helpers

import "github.com/gin-gonic/gin"

// APIResponse is the standard JSON envelope for ALL API responses
type APIResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

func Success(c *gin.Context, statusCode int, message string, data any) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Error(c *gin.Context, statusCode int, errMsg string) {
	c.JSON(statusCode, APIResponse{
		Success: false,
		Error:   errMsg,
	})
}
