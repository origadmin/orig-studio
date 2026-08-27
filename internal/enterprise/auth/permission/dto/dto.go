package dto

import "time"

type GroupItem struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description,omitempty"`
	Permissions   []string  `json:"permissions"`
	CategoryScope []string  `json:"category_scope,omitempty"`
	IsActive      bool      `json:"is_active"`
	CreatedBy     string    `json:"created_by,omitempty"`
	MemberCount   int       `json:"member_count"`
	CreateTime    time.Time `json:"create_time"`
	UpdateTime    time.Time `json:"update_time"`
}

type MemberItem struct {
	ID       string    `json:"id"`
	UserID   string    `json:"user_id"`
	Username string    `json:"username,omitempty"`
	GroupID  string    `json:"group_id"`
	JoinedAt time.Time `json:"joined_at"`
}

type UserPermissionDetail struct {
	UserID               string            `json:"user_id"`
	Role                 string            `json:"role"`
	EffectivePermissions map[string]*Source `json:"effective_permissions"`
	Groups               []UserGroupInfo   `json:"groups"`
}

type Source struct {
	Sources []string `json:"sources"`
	Scope   []string `json:"scope,omitempty"`
}

type UserGroupInfo struct {
	ID       string    `json:"id"`
	Name     string    `json:"name"`
	IsActive bool      `json:"is_active"`
	JoinedAt time.Time `json:"joined_at"`
}