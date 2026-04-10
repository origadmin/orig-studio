package server

import (
	"github.com/gin-gonic/gin"
)

// Response 统一响应格式
type Response[T any] struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data,omitempty"`
}

// PageResponse 分页响应格式
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

// NotificationListResponse 通知列表响应格式
type NotificationListResponse struct {
	Items       []interface{} `json:"items"`
	Total       int64         `json:"total"`
	UnreadCount int64         `json:"unread_count"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
}

// OK 返回成功响应
func OK(c *gin.Context, data interface{}) {
	c.JSON(200, Response[interface{}]{Code: 0, Message: "ok", Data: data})
}

// Fail 返回失败响应
func Fail(c *gin.Context, code int, message string) {
	c.JSON(getHTTPStatus(code), Response[interface{}]{Code: code, Message: message})
}

// getHTTPStatus 根据错误码获取HTTP状态码
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
	default:
		return 500
	}
}
