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
