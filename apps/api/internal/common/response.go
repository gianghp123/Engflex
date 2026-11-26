package common

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Pagination carries page metadata for list endpoints.
type Pagination struct {
	Page       int   `json:"page"`
	PageSize   int   `json:"pageSize"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

// NewPagination builds Pagination metadata. Page and pageSize are expected
// to be 1-based, already sanitized values (see ParsePagination).
func NewPagination(page, pageSize int, total int64) *Pagination {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 1
	}
	totalPages := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &Pagination{
		Page:       page,
		PageSize:   pageSize,
		Total:      total,
		TotalPages: totalPages,
	}
}

// ApiResponse is the shared JSON envelope for every API response.
// Pagination is nil for non-paginated responses.
type ApiResponse struct {
	Message    string      `json:"message"`
	Data       any         `json:"data,omitempty"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

// Success writes a non-paginated success response.
func Success(c *gin.Context, status int, message string, data any) {
	c.JSON(status, ApiResponse{
		Message: message,
		Data:    data,
	})
}

// OK writes a 200 response.
func OK(c *gin.Context, message string, data any) {
	Success(c, http.StatusOK, message, data)
}

// Created writes a 201 response.
func Created(c *gin.Context, message string, data any) {
	Success(c, http.StatusCreated, message, data)
}

// Paginated writes a success response with pagination metadata.
func Paginated(c *gin.Context, message string, data any, page, pageSize int, total int64) {
	c.JSON(http.StatusOK, ApiResponse{
		Message:    message,
		Data:       data,
		Pagination: NewPagination(page, pageSize, total),
	})
}

// Fail writes an *AppError as a failure envelope. Any other error becomes a
// 500 with its message hidden behind the default status text.
func Fail(c *gin.Context, err error) {
	appErr, ok := err.(*AppError)
	if !ok || appErr == nil {
		appErr = Internal()
	}
	resp := ApiResponse{Message: appErr.Message}
	if appErr.Details != nil {
		resp.Data = appErr.Details
	}
	c.JSON(appErr.Status, resp)
}
