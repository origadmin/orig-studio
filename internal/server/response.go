package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"

	http2 "origadmin/application/origcms/internal/helpers/http"
	ginadapter "origadmin/application/origcms/internal/helpers/http/gin"
	"origadmin/application/origcms/internal/infra/auth"
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

// PageData is the standard pagination data wrapper used by Page.
type PageData struct {
	Items    interface{} `json:"items"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// protojsonMarshaler is the shared protojson marshaler for consistent serialization.
var protojsonMarshaler = protojson.MarshalOptions{
	EmitUnpopulated: true,
	UseProtoNames:   true,
}

// writeProtoResponse writes a proto.Message as a unified JSON response with the given HTTP status.
func writeProtoResponse(c *gin.Context, statusCode int, data proto.Message) {
	b, err := protojsonMarshaler.Marshal(data)
	if err != nil {
		Fail(c, ErrInternal, "internal error")
		return
	}
	resp := fmt.Sprintf(`{"code":0,"message":"ok","data":%s}`, string(b))
	c.Data(statusCode, "application/json; charset=utf-8", []byte(resp))
}

// writeJSONResponse writes a non-proto data as a unified JSON response with the given HTTP status.
func writeJSONResponse(c *gin.Context, statusCode int, data interface{}) {
	c.JSON(statusCode, Response[interface{}]{Code: 0, Message: "ok", Data: data})
}

// OK returns a success response with HTTP 200.
// It automatically detects the data type: if data implements proto.Message,
// protojson serialization is used (ensuring correct field names and timestamp formats);
// otherwise, standard encoding/json is used via Gin's JSON renderer.
func OK(c *gin.Context, data interface{}) {
	if msg, ok := data.(proto.Message); ok {
		writeProtoResponse(c, http.StatusOK, msg)
	} else {
		writeJSONResponse(c, http.StatusOK, data)
	}
}

// Page returns a paginated success response with HTTP 200.
// Like OK, it automatically detects proto.Message for correct serialization.
// For proto messages (typically ListXxxResponse), pagination metadata is part of
// the Proto Response message itself, so the data is serialized directly.
// For non-proto data, the items are wrapped in a standard PageData structure.
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

// Created returns a success response with HTTP 201 for resource creation.
// It automatically detects proto.Message for correct serialization.
func Created(c *gin.Context, data interface{}) {
	if msg, ok := data.(proto.Message); ok {
		writeProtoResponse(c, http.StatusCreated, msg)
	} else {
		c.JSON(http.StatusCreated, Response[interface{}]{Code: 0, Message: "ok", Data: data})
	}
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

// ---------------------------------------------------------------------------
// Deprecated aliases — kept for backward compatibility, will be removed later.
// ---------------------------------------------------------------------------

// ProtoOK is a deprecated alias for OK.
// OK now auto-detects proto.Message, so ProtoOK is no longer needed.
//
// Deprecated: Use OK instead.
func ProtoOK(c *gin.Context, data proto.Message) {
	writeProtoResponse(c, http.StatusOK, data)
}

// ProtoOKPage is a deprecated alias for Page.
// Page now auto-detects proto.Message, so ProtoOKPage is no longer needed.
//
// Deprecated: Use Page instead.
func ProtoOKPage(c *gin.Context, data proto.Message) {
	writeProtoResponse(c, http.StatusOK, data)
}

// ProtoCreated is a deprecated alias for Created.
// Created now auto-detects proto.Message, so ProtoCreated is no longer needed.
//
// Deprecated: Use Created instead.
func ProtoCreated(c *gin.Context, data proto.Message) {
	writeProtoResponse(c, http.StatusCreated, data)
}

// OKPage is a deprecated alias for Page.
//
// Deprecated: Use Page instead.
func OKPage(c *gin.Context, items interface{}, total int64, page, pageSize int) {
	Page(c, items, total, page, pageSize)
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

// ==================== http2.Context bridge functions ====================

// GetClaimsCtx retrieves claims from an http2.Context.
// It uses ctx.Get("claims") to access the claims set by middleware.
func GetClaimsCtx(ctx http2.Context) (*auth.Claims, bool) {
	val, ok := ctx.Get("claims")
	if !ok {
		return nil, false
	}
	claims, ok := val.(*auth.Claims)
	return claims, ok
}

// HTTPToHandlerFunc wraps a standard http.HandlerFunc into an http2.HandlerFunc.
// It extracts the gin.Context from the http2.Context and passes it through
// the request context so that ginadapter.GetGinContext(r) still works inside
// the handler. This is a bridge function for the migration period.
func HTTPToHandlerFunc(h http.HandlerFunc) http2.HandlerFunc {
	return func(ctx http2.Context) error {
		gc := ginadapter.GinContextFromHTTP(ctx)
		if gc == nil {
			http2.Fail(ctx, http2.ErrInternal, "internal error")
			return nil
		}
		r := ginadapter.SetGinContext(ctx.Request(), gc)
		h(gc.Writer, r)
		return nil
	}
}

// GinHandlerToHandlerFunc wraps a gin handler function (func(*gin.Context))
// into an http2.HandlerFunc. This is a bridge function for handlers that
// directly accept *gin.Context as a parameter during the migration period.
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
