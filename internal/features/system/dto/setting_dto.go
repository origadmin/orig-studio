/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

package dto

import "time"

// SettingType mirrors entity/setting.Type without importing the entity package.
type SettingType string

const (
	SettingTypeBool  SettingType = "bool"
	SettingTypeInt   SettingType = "int"
	SettingTypeString SettingType = "string"
	SettingTypeJSON  SettingType = "json"
)

// SettingCategory mirrors entity/setting.Category without importing the entity package.
type SettingCategory string

const (
	SettingCategoryGeneral   SettingCategory = "general"
	SettingCategoryStorage   SettingCategory = "storage"
	SettingCategoryEmail     SettingCategory = "email"
	SettingCategorySecurity  SettingCategory = "security"
	SettingCategoryAdvanced  SettingCategory = "advanced"
	SettingCategoryPortal    SettingCategory = "portal"
)

// SettingDTO is the data transfer object for settings, isolating biz/service from entity.
type SettingDTO struct {
	ID            string         `json:"id,omitempty"`
	Key           string         `json:"key,omitempty"`
	Value         string         `json:"value,omitempty"`
	Type          SettingType    `json:"type,omitempty"`
	Category      SettingCategory `json:"category,omitempty"`
	Description   string         `json:"description,omitempty"`
	IsSensitive   bool           `json:"is_sensitive,omitempty"`
	FallbackValue string         `json:"fallback_value,omitempty"`
	IsBuiltin     bool           `json:"is_builtin,omitempty"`
	CreateTime    time.Time      `json:"create_time,omitempty"`
	UpdateTime    time.Time      `json:"update_time,omitempty"`
}
