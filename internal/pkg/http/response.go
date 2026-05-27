package http

import (
	"net/http"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

var protojsonMarshaler = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

const (
	ErrOK           = 0
	ErrBadRequest   = 40000
	ErrUnauthorized = 40100
	ErrForbidden    = 40300
	ErrNotFound     = 40400
	ErrConflict     = 40900
	ErrInternal     = 50000

	AppErrInternal     = 10000
	AppErrNotFound     = 10001
	AppErrUnauthorized = 10002
	AppErrForbidden    = 10003
	AppErrBadRequest   = 10004
	AppErrConflict     = 10005

	AppErrUserNotFound  = 20001
	AppErrUserExists    = 20002
	AppErrPasswordWrong = 20003
	AppErrTokenExpired  = 20004
	AppErrTokenInvalid  = 20005

	AppErrMediaNotFound  = 30001
	AppErrMediaTooLarge  = 30002
	AppErrMediaForbidden = 30003
	AppErrEncodingFailed = 30004

	AppErrCommentNotFound  = 40001
	AppErrCommentForbidden = 40002
)

type ErrorResponse struct {
	Code     int               `json:"code"`
	Reason   string            `json:"reason"`
	Message  string            `json:"message"`
	Metadata map[string]string `json:"metadata"`
}

type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func OK(ctx Context, data interface{}) error {
	if msg, ok := data.(proto.Message); ok {
		return writeProtoResponse(ctx, http.StatusOK, msg)
	}
	return ctx.Result(http.StatusOK, data)
}

func Created(ctx Context, data interface{}) error {
	if msg, ok := data.(proto.Message); ok {
		return writeProtoResponse(ctx, http.StatusCreated, msg)
	}
	return ctx.Result(http.StatusCreated, data)
}

func Fail(ctx Context, code int, message string) error {
	status := errorToHTTPStatus(code)
	reason := errorCodeToReason(code)
	return ctx.Result(status, ErrorResponse{Code: status, Reason: reason, Message: message, Metadata: map[string]string{}})
}

func Page(ctx Context, items interface{}, total int64, page, pageSize int) error {
	data := &PageData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}
	return ctx.Result(http.StatusOK, data)
}

func writeProtoResponse(ctx Context, statusCode int, data proto.Message) error {
	b, err := protojsonMarshaler.Marshal(data)
	if err != nil {
		return err
	}
	ctx.Response().Header().Set("Content-Type", "application/json; charset=utf-8")
	ctx.Response().WriteHeader(statusCode)
	_, err = ctx.Response().Write(b)
	return err
}

func errorCodeToReason(code int) string {
	switch code {
	case ErrNotFound:
		return "NOT_FOUND"
	case AppErrNotFound:
		return "NOT_FOUND"
	case AppErrUserNotFound:
		return "USER_NOT_FOUND"
	case AppErrMediaNotFound:
		return "MEDIA_NOT_FOUND"
	case AppErrCommentNotFound:
		return "COMMENT_NOT_FOUND"
	case ErrUnauthorized:
		return "UNAUTHORIZED"
	case AppErrUnauthorized:
		return "UNAUTHORIZED"
	case AppErrPasswordWrong:
		return "PASSWORD_WRONG"
	case AppErrTokenExpired:
		return "TOKEN_EXPIRED"
	case AppErrTokenInvalid:
		return "TOKEN_INVALID"
	case ErrForbidden:
		return "FORBIDDEN"
	case AppErrForbidden:
		return "FORBIDDEN"
	case AppErrMediaForbidden:
		return "MEDIA_FORBIDDEN"
	case AppErrCommentForbidden:
		return "COMMENT_FORBIDDEN"
	case ErrBadRequest:
		return "BAD_REQUEST"
	case AppErrBadRequest:
		return "BAD_REQUEST"
	case ErrConflict:
		return "CONFLICT"
	case AppErrConflict:
		return "CONFLICT"
	case AppErrUserExists:
		return "USER_EXISTS"
	case AppErrMediaTooLarge:
		return "PAYLOAD_TOO_LARGE"
	case AppErrEncodingFailed:
		return "ENCODING_FAILED"
	case ErrInternal:
		return "INTERNAL_ERROR"
	case AppErrInternal:
		return "INTERNAL_ERROR"
	default:
		if code >= 10000 {
			return "INTERNAL_ERROR"
		}
		return "ERROR"
	}
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
