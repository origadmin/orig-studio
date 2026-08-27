/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RequestIDHeader is the header key for the request ID.
const RequestIDHeader = "X-Request-ID"

// RequestIDContextKey is the context key type for request ID values.
type RequestIDContextKey struct{}

// RequestID returns a gin middleware that extracts or generates a request ID
// and propagates it via the gin context, response header, and request context.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = uuid.NewString()
		}

		// Set in gin context for handlers using c.GetString("request_id")
		c.Set("request_id", requestID)

		// Set in response header so clients can correlate
		c.Header(RequestIDHeader, requestID)

		// Inject into request context for downstream code using context.Value
		ctx := context.WithValue(c.Request.Context(), RequestIDContextKey{}, requestID)
		c.Request = c.Request.WithContext(ctx)

		c.Next()
	}
}

// RequestIDFromContext extracts the request ID from a context.
// Returns empty string if not set.
func RequestIDFromContext(ctx context.Context) string {
	if v, ok := ctx.Value(RequestIDContextKey{}).(string); ok {
		return v
	}
	return ""
}
