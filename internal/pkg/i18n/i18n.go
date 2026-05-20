/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package i18n provides internationalization utilities.
package i18n

// Text returns a translation key. Currently returns the key itself.
// In production, this would look up the translation from a localization system.
func Text(key string) string {
	return key
}
