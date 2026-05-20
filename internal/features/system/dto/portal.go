/*
 * Copyright (c) 2024 OrigAdmin. All rights reserved.
 */

// Package dto provides data transfer objects for the system feature module.
package dto

// PortalModuleConfig is the DTO for portal module configuration.
type PortalModuleConfig struct {
	Articles bool `json:"articles"`
	Videos   bool `json:"videos"`
	Music    bool `json:"music"`
}

// PortalSiteConfig is the DTO for portal site configuration.
type PortalSiteConfig struct {
	SiteName          string   `json:"site_name"`
	SiteDescription   string   `json:"site_description"`
	PrimaryURL        string   `json:"primary_url"`
	AllowedURLs       []string `json:"allowed_urls"`
	AllowRegistration bool     `json:"allow_registration"`
	AllowUpload       bool     `json:"allow_upload"`
}

// PortalConfigResponse is the DTO for portal configuration responses.
type PortalConfigResponse struct {
	Modules PortalModuleConfig `json:"modules"`
	Layout  string             `json:"layout"`
	Site    PortalSiteConfig   `json:"site"`
}
