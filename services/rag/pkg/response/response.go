package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIResponse is the standard envelope for all API responses.
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Success sends a 200 OK response with code 0 and the provided data.
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Created sends a 201 Created response with code 0 and the provided data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// Error sends an error response with the given HTTP status code and message.
func Error(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, APIResponse{
		Code:    httpCode,
		Message: message,
	})
}
