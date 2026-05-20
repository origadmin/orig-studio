/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package types

import "errors"

// IsNotFound checks if the given error is a "not found" error.
// This is a domain-level wrapper that avoids importing internal/dal/entity
// in biz/ and service/ layers.
// The dal/ layer provides the concrete implementation via RegisterNotFoundChecker.
var notFoundChecker func(error) bool

// RegisterNotFoundChecker registers the concrete not-found checker from the dal/entity layer.
// This should be called once during application initialization (before any biz/service calls).
func RegisterNotFoundChecker(checker func(error) bool) {
	notFoundChecker = checker
}

// IsNotFound returns true if the error represents a "not found" condition.
// Falls back to checking for Go standard errors.Is(err, ErrNotFound) if no
// checker has been registered.
func IsNotFound(err error) bool {
	if err == nil {
		return false
	}
	if notFoundChecker != nil {
		return notFoundChecker(err)
	}
	// Fallback: check for common not-found patterns
	return errors.Is(err, ErrNotFound)
}

// ErrNotFound is a sentinel error for not-found conditions.
var ErrNotFound = errors.New("not found")
