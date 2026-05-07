package http

import (
	"fmt"
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protojsonMarshaler = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

const (
	// HTTP-style error codes (aligned with HTTP status code * 100)
	ErrOK           = 0
	ErrBadRequest   = 40000
	ErrUnauthorized = 40100
	ErrForbidden    = 40300
	ErrNotFound     = 40400
	ErrConflict     = 40900
	ErrInternal     = 50000

	// Application error codes (aligned with server/error.go)
	// Common errors (1xxxx)
	AppErrInternal     = 10000
	AppErrNotFound     = 10001
	AppErrUnauthorized = 10002
	AppErrForbidden    = 10003
	AppErrBadRequest   = 10004
	AppErrConflict     = 10005

	// User errors (2xxxx)
	AppErrUserNotFound  = 20001
	AppErrUserExists    = 20002
	AppErrPasswordWrong = 20003
	AppErrTokenExpired  = 20004
	AppErrTokenInvalid  = 20005

	// Media errors (3xxxx)
	AppErrMediaNotFound  = 30001
	AppErrMediaTooLarge  = 30002
	AppErrMediaForbidden = 30003
	AppErrEncodingFailed = 30004

	// Comment errors (4xxxx)
	AppErrCommentNotFound  = 40001
	AppErrCommentForbidden = 40002
)

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func OK(ctx Context, data interface{}) error {
	return writeProtoOrJSON(ctx, http.StatusOK, data)
}

func Created(ctx Context, data interface{}) error {
	return writeProtoOrJSON(ctx, http.StatusCreated, data)
}

func Fail(ctx Context, code int, message string) error {
	status := errorToHTTPStatus(code)
	resp := &Response{Code: code, Message: message}
	return ctx.Result(status, resp)
}

func Page(ctx Context, items interface{}, total int64, page, pageSize int) error {
	data := &PageData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	return writeProtoOrJSON(ctx, http.StatusOK, data)
}

func writeProtoOrJSON(ctx Context, statusCode int, data interface{}) error {
	if msg, ok := data.(proto.Message); ok {
		b, err := protojsonMarshaler.Marshal(msg)
		if err != nil {
			return err
		}
		wrapped := fmt.Sprintf(`{"code":0,"message":"ok","data":%s}`, string(b))
		ctx.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
		ctx.Response().WriteHeader(statusCode)
		_, err = ctx.Response().Write([]byte(wrapped))
		return err
	}
	return ctx.Result(statusCode, data)
}

func errorToHTTPStatus(code int) int {
	switch code {
	case ErrOK, AppErrInternal:
		return http.StatusOK
	case ErrNotFound, AppErrNotFound, AppErrUserNotFound, AppErrMediaNotFound, AppErrCommentNotFound:
		return http.StatusNotFound
	case ErrUnauthorized, AppErrUnauthorized, AppErrPasswordWrong, AppErrTokenExpired, AppErrTokenInvalid:
		return http.StatusUnauthorized
	case ErrForbidden, AppErrForbidden, AppErrMediaForbidden, AppErrCommentForbidden:
		return http.StatusForbidden
	case ErrBadRequest, AppErrBadRequest:
		return http.StatusBadRequest
	case ErrConflict, AppErrConflict, AppErrUserExists:
		return http.StatusConflict
	case AppErrMediaTooLarge:
		return http.StatusRequestEntityTooLarge
	case AppErrEncodingFailed:
		return http.StatusInternalServerError
	default:
		if code >= 10000 {
			return http.StatusInternalServerError
		}
		return http.StatusOK
	}
}
