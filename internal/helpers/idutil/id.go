/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package idutil provides ID utility functions.
package idutil

import "fmt"

// FormatUserID formats a user ID as a string.
func FormatUserID(id int64) string {
	return fmt.Sprintf("%d", id)
}
