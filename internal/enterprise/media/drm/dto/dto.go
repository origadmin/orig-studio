package dto

import "time"

type DrmPolicyDTO struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Type            string    `json:"type"`
	HlsKeyURL       string    `json:"hls_key_url,omitempty"`
	WidevinePssh    string    `json:"widevine_pssh,omitempty"`
	FairplayCertURL string    `json:"fairplay_cert_url,omitempty"`
	IsDefault       bool      `json:"is_default"`
	Description     string    `json:"description,omitempty"`
	CreateTime      time.Time `json:"create_time,omitempty"`
	UpdateTime      time.Time `json:"update_time,omitempty"`
}

type DrmKeyDTO struct {
	ID        string    `json:"id"`
	PolicyID  string    `json:"policy_id"`
	ContentID string    `json:"content_id"`
	KeyID     string    `json:"key_id"`
	Iv        string    `json:"iv,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}

type DrmLicenseDTO struct {
	ID        string    `json:"id"`
	KeyID     string    `json:"key_id"`
	UserID    string    `json:"user_id,omitempty"`
	DeviceID  string    `json:"device_id,omitempty"`
	Status    string    `json:"status"`
	IssuedAt  time.Time `json:"issued_at,omitempty"`
	ExpiresAt time.Time `json:"expires_at,omitempty"`
}