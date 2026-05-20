/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 * UserEntityDTO provides entity-level fields not available in proto types (e.g., Role).
 * This isolates biz/ and service/ layers from internal/dal/entity imports.
 */

package dto

import "time"

// UserRoleType mirrors entity/user.Role without importing the entity package.
type UserRoleType string

const (
	UserRoleUser   UserRoleType = "user"
	UserRoleAdmin  UserRoleType = "admin"
	UserRoleEditor UserRoleType = "editor"
)

// UserEntityDTO carries entity-level user fields (e.g., Role) that are not
// present in the proto-generated types.User. The biz/ and service/ layers
// use this instead of importing entity.User directly.
type UserEntityDTO struct {
	ID            string       `json:"id,omitempty"`
	Username      string       `json:"username,omitempty"`
	Email         string       `json:"email,omitempty"`
	Name          string       `json:"name,omitempty"`
	Slug          string       `json:"slug,omitempty"`
	Role          UserRoleType `json:"role,omitempty"`
	IsStaff       bool         `json:"is_staff,omitempty"`
	IsSuperuser   bool         `json:"is_superuser,omitempty"`
	IsFeatured    bool         `json:"is_featured,omitempty"`
	IsEditor      bool         `json:"is_editor,omitempty"`
	AdvancedUser  bool         `json:"advanced_user,omitempty"`
	Logo          string       `json:"logo,omitempty"`
	Status        string       `json:"status,omitempty"`
	DateJoined    time.Time    `json:"date_joined,omitempty"`
	CreateTime    time.Time    `json:"create_time,omitempty"`
	UpdateTime    time.Time    `json:"update_time,omitempty"`
}
