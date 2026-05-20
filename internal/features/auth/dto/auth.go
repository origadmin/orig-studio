/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package dto provides data transfer objects for the auth feature module.
package dto

// LoginRequest is the DTO for login requests.
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// RegisterRequest is the DTO for registration requests.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=64"`
	Password string `json:"password" binding:"required,min=6,max=128"`
	Email    string `json:"email"    binding:"omitempty,email"`
	Nickname string `json:"nickname"`
}

// TokenResponse is the DTO for authentication token responses.
type TokenResponse struct {
	AccessToken  string     `json:"access_token"`
	RefreshToken string     `json:"refresh_token"`
	TokenType    string     `json:"token_type"`
	ExpiresIn    int64      `json:"expires_in"`
	User         *LoginUser `json:"user"`
}

// LoginUser is the DTO for user info in auth responses.
type LoginUser struct {
	Id       string `json:"id"`
	Username string `json:"username"`
	Nickname string `json:"nickname,omitempty"`
	Email    string `json:"email,omitempty"`
	IsStaff  bool   `json:"is_staff"`
}
