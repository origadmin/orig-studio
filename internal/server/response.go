package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
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

// PageData is the standard pagination data wrapper used by OKPage.
type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// OK returns a success response with HTTP 200.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response[interface{}]{Code: 0, Message: "ok", Data: data})
}

// OKPage returns a paginated success response with HTTP 200.
func OKPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	c.JSON(http.StatusOK, Response[interface{}]{
		Code:    0,
		Message: "ok",
		Data: PageData{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		},
	})
}

// Created returns a success response with HTTP 201 for resource creation.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response[interface{}]{Code: 0, Message: "ok", Data: data})
}

// Fail returns a failure response with the appropriate HTTP status code.
func Fail(c *gin.Context, code int, message string) {
	c.JSON(getHTTPStatus(code), Response[interface{}]{Code: code, Message: message})
}

// FailAbort returns a failure response and aborts the middleware chain.
// Use this in middleware that needs to stop further processing.
func FailAbort(c *gin.Context, code int, message string) {
	c.AbortWithStatusJSON(getHTTPStatus(code), Response[interface{}]{Code: code, Message: message})
}

// protojsonMarshaler is the shared protojson marshaler for consistent serialization.
var protojsonMarshaler = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

// ProtoOK returns a success response with protojson serialization for proto.Message types.
// This ensures timestamppb and other proto types are serialized correctly.
func ProtoOK(c *gin.Context, data proto.Message) {
	b, err := protojsonMarshaler.Marshal(data)
	if err != nil {
		Fail(c, ErrInternal, "internal error")
		return
	}
	resp := fmt.Sprintf(`{"code":0,"message":"ok","data":%s}`, string(b))
	c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(resp))
}

// getHTTPStatus maps application error codes to HTTP status codes.
func getHTTPStatus(code int) int {
	switch {
	case code == 0:
		return http.StatusOK
	case code == ErrNotFound:
		return http.StatusNotFound
	case code == ErrUnauthorized:
		return http.StatusUnauthorized
	case code == ErrForbidden:
		return http.StatusForbidden
	case code == ErrBadRequest:
		return http.StatusBadRequest
	case code == ErrConflict:
		return http.StatusConflict
	case code == ErrUserNotFound:
		return http.StatusNotFound
	case code == ErrUserExists:
		return http.StatusConflict
	case code == ErrPasswordWrong:
		return http.StatusUnauthorized
	case code == ErrTokenExpired:
		return http.StatusUnauthorized
	case code == ErrTokenInvalid:
		return http.StatusUnauthorized
	case code == ErrMediaNotFound:
		return http.StatusNotFound
	case code == ErrMediaTooLarge:
		return http.StatusRequestEntityTooLarge
	case code == ErrMediaForbidden:
		return http.StatusForbidden
	case code == ErrEncodingFailed:
		return http.StatusInternalServerError
	case code == ErrCommentNotFound:
		return http.StatusNotFound
	case code == ErrCommentForbidden:
		return http.StatusForbidden
	default:
		if code >= 10000 {
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}
}
