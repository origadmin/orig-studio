package server

import (
	"github.com/gin-gonic/gin"
)

// Response unified response format
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

// PageResponse pagination response format
type PageResponse[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items    []T   `json:"items"`
		Total    int64 `json:"total"`
		Page     int   `json:"page"`
		PageSize int   `json:"page_size"`
	} `json:"data"`
}

// NotificationListResponse notification list response format
type NotificationListResponse struct {
	Items       []interface{} `json:"items"`
	Total       int64         `json:"total"`
	UnreadCount int64         `json:"unread_count"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
}

// OK return success response
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Response[interface{}]{Code: 0, Message: "ok", Data: data})
}

// Fail return failure response
func Fail(c *gin.Context, code int, message string) {
	c.JSON(getHTTPStatus(code), Response[interface{}]{Code: code, Message: message})
}

// getHTTPStatus get HTTP status code based on error code
func getHTTPStatus(code int) int {
	switch {
	case code == 0:
		return 200
	case code == 10001:
		return 404
	case code == 10002:
		return 401
	case code == 10003:
		return 403
	case code == 10004:
		return 400
	case code == 10005:
		return 409
	case code == 20001:
		return 404 // ErrUserNotFound
	case code == 30001:
		return 404 // ErrMediaNotFound
	case code == 40001:
		return 404 // ErrCommentNotFound
	default:
		return 500
	}
}
