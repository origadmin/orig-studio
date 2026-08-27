package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	http2 "origadmin/application/origstudio/internal/pkg/http"
	ginadapter "origadmin/application/origstudio/internal/pkg/http/gin"
	"origadmin/application/origstudio/internal/infra/auth"
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

type NotificationListResponse struct {
	Items       []interface{} `json:"items"`
	Total       int64         `json:"total"`
	UnreadCount int64         `json:"unread_count"`
	Page        int           `json:"page"`
	PageSize    int           `json:"page_size"`
}

var protojsonMarshaler = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

func writeProtoResponse(c *gin.Context, statusCode int, data proto.Message) {
	b, err := protojsonMarshaler.Marshal(data)
	if err != nil {
		Fail(c, ErrInternal, "internal error")
		return
	}
	c.Data(statusCode, "application/json; charset=utf-8", b)
}

func writeJSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, data)
}

func OK(c *gin.Context, data interface{}) {
	if msg, ok := data.(proto.Message); ok {
		writeProtoResponse(c, http.StatusOK, msg)
	} else {
		writeJSONResponse(c, http.StatusOK, data)
	}
}

func Page(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	if msg, ok := items.(proto.Message); ok {
		writeProtoResponse(c, http.StatusOK, msg)
	} else {
		writeJSONResponse(c, http.StatusOK, PageData{
			Items:    items,
			Total:    total,
			Page:     page,
			PageSize: pageSize,
		})
	}
}

func Created(c *gin.Context, data interface{}) {
	if msg, ok := data.(proto.Message); ok {
		writeProtoResponse(c, http.StatusCreated, msg)
	} else {
		writeJSONResponse(c, http.StatusCreated, data)
	}
}

func Fail(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	reason := errorCodeToReason(code)
	c.JSON(httpStatus, ErrorResponse{Code: httpStatus, Reason: reason, Message: message, Metadata: map[string]string{}})
}

func FailAbort(c *gin.Context, code int, message string) {
	httpStatus := getHTTPStatus(code)
	reason := errorCodeToReason(code)
	c.AbortWithStatusJSON(httpStatus, ErrorResponse{Code: httpStatus, Reason: reason, Message: message, Metadata: map[string]string{}})
}

func ProtoOK(c *gin.Context, data proto.Message) {
	writeProtoResponse(c, http.StatusOK, data)
}

func ProtoOKPage(c *gin.Context, data proto.Message) {
	writeProtoResponse(c, http.StatusOK, data)
}

func ProtoCreated(c *gin.Context, data proto.Message) {
	writeProtoResponse(c, http.StatusCreated, data)
}

func OKPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	Page(c, items, total, page, pageSize)
}

func errorCodeToReason(code int) string {
	switch code {
	case ErrNotFound:
		return "NOT_FOUND"
	case ErrUserNotFound:
		return "USER_NOT_FOUND"
	case ErrMediaNotFound:
		return "MEDIA_NOT_FOUND"
	case ErrCommentNotFound:
		return "COMMENT_NOT_FOUND"
	case ErrUnauthorized:
		return "UNAUTHORIZED"
	case ErrTokenExpired:
		return "TOKEN_EXPIRED"
	case ErrTokenInvalid:
		return "TOKEN_INVALID"
	case ErrPasswordWrong:
		return "PASSWORD_WRONG"
	case ErrForbidden:
		return "FORBIDDEN"
	case ErrMediaForbidden:
		return "MEDIA_FORBIDDEN"
	case ErrCommentForbidden:
		return "COMMENT_FORBIDDEN"
	case ErrBadRequest:
		return "BAD_REQUEST"
	case ErrPayloadTooLarge:
		return "PAYLOAD_TOO_LARGE"
	case ErrUnsupportedMediaType:
		return "UNSUPPORTED_MEDIA_TYPE"
	case ErrConflict:
		return "CONFLICT"
	case ErrUserExists:
		return "USER_EXISTS"
	case ErrMediaTooLarge:
		return "PAYLOAD_TOO_LARGE"
	case ErrEncodingFailed:
		return "ENCODING_FAILED"
	case ErrInternal:
		return "INTERNAL_ERROR"
	default:
		if code >= 10000 {
			return "INTERNAL_ERROR"
		}
		return "ERROR"
	}
}

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
	case code == ErrPayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case code == ErrUnsupportedMediaType:
		return http.StatusUnsupportedMediaType
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

func GetClaimsCtx(ctx http2.Context) (*auth.Claims, bool) {
	val, ok := ctx.Get("claims")
	if !ok {
		return nil, false
	}
	claims, ok := val.(*auth.Claims)
	return claims, ok
}

// ==================== http2.Context variants ====================
// These functions accept the framework-agnostic http2.Context and are used by
// service handlers that have been migrated off *gin.Context. The legacy
// *gin.Context variants (OK/Fail/Page/Created/ProtoOK/FailAbort) are retained
// for CE compatibility and existing tests; new code should use the Ctx variants.
//
// Behavior parity: proto.Message → protojsonMarshaler (snake_case + EmitUnpopulated);
// non-proto → encoding/json via ctx.JSON. This matches the *gin.Context path so
// CE and EE produce identical responses.

func writeCtxProtoResponse(ctx http2.Context, statusCode int, data proto.Message) error {
	b, err := protojsonMarshaler.Marshal(data)
	if err != nil {
		return FailCtx(ctx, ErrInternal, "internal error")
	}
	return ctx.Blob(statusCode, "application/json; charset=utf-8", b)
}

func writeCtxJSONResponse(ctx http2.Context, statusCode int, data interface{}) error {
	return ctx.JSON(statusCode, data)
}

func OKCtx(ctx http2.Context, data interface{}) error {
	if msg, ok := data.(proto.Message); ok {
		return writeCtxProtoResponse(ctx, http.StatusOK, msg)
	}
	return writeCtxJSONResponse(ctx, http.StatusOK, data)
}

func PageCtx(ctx http2.Context, items interface{}, total int64, page, pageSize int) {
	if msg, ok := items.(proto.Message); ok {
		_ = writeCtxProtoResponse(ctx, http.StatusOK, msg)
		return
	}
	_ = writeCtxJSONResponse(ctx, http.StatusOK, PageData{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func CreatedCtx(ctx http2.Context, data interface{}) error {
	if msg, ok := data.(proto.Message); ok {
		return writeCtxProtoResponse(ctx, http.StatusCreated, msg)
	}
	return writeCtxJSONResponse(ctx, http.StatusCreated, data)
}

func FailCtx(ctx http2.Context, code int, message string) error {
	httpStatus := getHTTPStatus(code)
	reason := errorCodeToReason(code)
	return ctx.JSON(httpStatus, ErrorResponse{Code: httpStatus, Reason: reason, Message: message, Metadata: map[string]string{}})
}

func FailAbortCtx(ctx http2.Context, code int, message string) error {
	// http2.Context has no Abort concept; FailAbortCtx behaves like FailCtx.
	return FailCtx(ctx, code, message)
}

func ProtoOKCtx(ctx http2.Context, data proto.Message) error {
	return writeCtxProtoResponse(ctx, http.StatusOK, data)
}

func ProtoOKPageCtx(ctx http2.Context, data proto.Message) error {
	return writeCtxProtoResponse(ctx, http.StatusOK, data)
}

func ProtoCreatedCtx(ctx http2.Context, data proto.Message) error {
	return writeCtxProtoResponse(ctx, http.StatusCreated, data)
}

func OKPageCtx(ctx http2.Context, items interface{}, total int64, page, pageSize int) {
	PageCtx(ctx, items, total, page, pageSize)
}

func HTTPToHandlerFunc(h http.HandlerFunc) http2.HandlerFunc {
	return func(ctx http2.Context) error {
		h(ctx.Response(), ctx.Request())
		return nil
	}
}

func GinHandlerToHandlerFunc(h func(*gin.Context)) http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		if gc == nil {
			http2.Fail(ctx, http2.ErrInternal, "internal error")
			return nil
		}
		h(gc)
		return nil
	}
}
