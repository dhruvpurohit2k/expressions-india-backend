package utils

import (
	"log"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Success bool       `json:"success"`
	Data    any        `json:"data,omitempty"`
	Error   *ErrorInfo `json:"error,omitempty"`
	Meta    *Meta      `json:"meta,omitempty"`
}

type ErrorInfo struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Meta struct {
	Page       int   `json:"page,omitempty"`
	PerPage    int   `json:"perPage,omitempty"`
	Total      int64 `json:"total,omitempty"`
	TotalPages int   `json:"totalPages,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
	})
}

func PaginatedOK(c *gin.Context, data any, meta Meta) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Data:    data,
		Meta:    &meta,
	})
}

// Send a error msg to the user.
//
// Takes in gin.Context, status code and message.
func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, Response{
		Success: false,
		Error:   &ErrorInfo{Code: code, Message: message},
	})
}

// FailInternal logs the full error server-side and returns a generic message to the client.
// Use this instead of Fail for 5xx errors to avoid leaking internal details.
func FailInternal(c *gin.Context, code string, clientMsg string, err error) {
	log.Printf("[%s] %s: %v", code, clientMsg, err)
	Fail(c, http.StatusInternalServerError, code, clientMsg)
}

// SafeTotalPages calculates total pages without dividing by zero.
func SafeTotalPages(total int64, limit int) int {
	if limit <= 0 {
		return 0
	}
	return int(math.Ceil(float64(total) / float64(limit)))
}
