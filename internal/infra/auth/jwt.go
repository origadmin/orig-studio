/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

func (c *Claims) GetUserID() string {
	return c.Subject
}

func (c *Claims) IsAdmin() bool {
	return c.Role == "admin"
}

type Manager struct {
	secret          []byte
	ttl             time.Duration
	refreshTokenTTL time.Duration
}

func NewManager(secret string, ttl time.Duration, refreshTokenTTL time.Duration) *Manager {
	return &Manager{
		secret:          []byte(secret),
		ttl:             ttl,
		refreshTokenTTL: refreshTokenTTL,
	}
}

func (m *Manager) TTL() time.Duration {
	return m.ttl
}

func (m *Manager) Generate(
	userID string,
	username string,
	role string,
) (string, error) {
	now := time.Now()
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.ttl)),
			Issuer:    "origcms",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

func (m *Manager) GenerateRefreshToken(
	userID string,
	username string,
	role string,
) (string, error) {
	now := time.Now()
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.refreshTokenTTL)),
			Issuer:    "origcms",
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

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
