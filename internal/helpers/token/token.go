package token

import (
	"crypto/rand"
	"math/big"
	"strings"
)

const (
	// TokenLength is the length of the short token
	TokenLength = 12
	// ShortIDLength is the length of the short ID
	ShortIDLength = 8
	// CharSet is the character set used for generating tokens
	CharSet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

// GenerateShortToken generates a random short token for URL references
func GenerateShortToken() (string, error) {
	return generateRandomString(TokenLength)
}

// GenerateShortID generates a random short ID for sharing
func GenerateShortID() (string, error) {
	return generateRandomString(ShortIDLength)
}

// generateRandomString generates a random string of the specified length
func generateRandomString(length int) (string, error) {
	var sb strings.Builder
	sb.Grow(length)

	charSetLength := big.NewInt(int64(len(CharSet)))

	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, charSetLength)
		if err != nil {
			return "", err
		}
		sb.WriteByte(CharSet[num.Int64()])
	}

	return sb.String(), nil
}
