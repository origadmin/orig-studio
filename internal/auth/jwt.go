/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package auth provides JWT token generation and validation utilities.
package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the JWT claims structure for origcms.
type Claims struct {
	UserID   int64  `json:"uid"`
	Username string `json:"username"`
	IsStaff  bool   `json:"is_staff"`
	Role string `json:"role"`
	jwt.RegisteredClaims
}

// Manager handles JWT signing and parsing.
type Manager struct {
	secret []byte
	ttl    time.Duration
}

// NewManager creates a new JWT Manager.
func NewManager(secret string, ttl time.Duration) *Manager {
	return &Manager{secret: []byte(secret), ttl: ttl}
}

// TTL returns the token time-to-live duration.
func (m *Manager) TTL() time.Duration {
	return m.ttl
}

// Generate creates a signed JWT token for the given user.
func (m *Manager) Generate(userID int64, username string, isStaff bool, role string) (string, error) {
	now := time.Now()
	claims := Claims{
		UserID:   userID,
		Username: username,
		IsStaff:  isStaff,
		Role: role,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			Issuer:    "origcms",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// Parse validates and parses a JWT token string.
func (m *Manager) Parse(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return m.secret, nil
	})
	if err != nil {
		return nil, err
	}
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}
	return claims, nil
}
