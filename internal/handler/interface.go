/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package handler defines transport-agnostic handler interfaces for the application.
package handler

import (
	"context"
	"net/http"
)

// Handler defines the interface for all transport-agnostic handlers.
type Handler interface {
	// Register registers the handler's routes.
	Register(r Router)
}

// Router defines the interface for a router that handlers can register with.
type Router interface {
	// Group creates a new route group with the given prefix.
	Group(prefix string) Router
	// GET registers a GET route.
	GET(path string, handler http.HandlerFunc)
	// POST registers a POST route.
	POST(path string, handler http.HandlerFunc)
	// PUT registers a PUT route.
	PUT(path string, handler http.HandlerFunc)
	// DELETE registers a DELETE route.
	DELETE(path string, handler http.HandlerFunc)
	// PATCH registers a PATCH route.
	PATCH(path string, handler http.HandlerFunc)
}

// Context defines the interface for a request context.
type Context interface {
	// Context returns the underlying context.Context.
	Context() context.Context
	// Request returns the underlying http.Request.
	Request() *http.Request
	// ResponseWriter returns the underlying http.ResponseWriter.
	ResponseWriter() http.ResponseWriter
	// Param gets a URL parameter.
	Param(key string) string
	// Query gets a query parameter.
	Query(key string) string
	// Bind binds the request body to a struct.
	Bind(v interface{}) error
	// JSON sends a JSON response.
	JSON(code int, v interface{})
	// Status sends a status code response.
	Status(code int)
	// Header sets a response header.
	Header(key, value string)
	// Get gets a value from the context.
	Get(key string) interface{}
	// Set sets a value in the context.
	Set(key string, value interface{})
}
