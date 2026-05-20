/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package idutil provides ID utility functions.
package idutil

import (
	"fmt"

	"github.com/origadmin/toolkits/identifier"
	_ "github.com/origadmin/toolkits/identifier/shortid"
	_ "github.com/origadmin/toolkits/identifier/snowflake"
	_ "github.com/origadmin/toolkits/identifier/uuid"
)

var (
	generator        = identifier.Get("snowflake")
	uuidGenerator    = identifier.Get("uuid-v7")
	shortIDGenerator = identifier.Get("shortid")
)

// FormatUserID formats a user ID as a string.
func FormatUserID(id int64) string {
	return fmt.Sprintf("%d", id)
}

// Gen generates a snowflake ID.
func Gen() int64 {
	v, _ := generator.GenerateNumber()
	return v
}

// GenUUID generates a UUID string.
func GenUUID() string {
	v, _ := uuidGenerator.GenerateString()
	return v
}

// GenStringUUID generates a UUID string and returns any error.
func GenStringUUID() (string, error) {
	v, err := uuidGenerator.GenerateString()
	return v, err
}

// GenUUIDv7 generates a UUIDv7 string.
func GenUUIDv7() string {
	return GenUUID()
}

// DefaultUUIDv7 returns a function that generates a UUIDv7 string.
func DefaultUUIDv7() func() string {
	return GenUUID
}

// GenShortID generates a short ID string.
func GenShortID() string {
	v, _ := shortIDGenerator.GenerateString()
	return v
}

// DefaultShortID returns a function that generates a short ID string.
func DefaultShortID() func() string {
	return GenShortID
}
