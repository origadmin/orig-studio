package dto

type TenantDTO struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Slug        string                 `json:"slug"`
	Domain      string                 `json:"domain,omitempty"`
	Logo        string                 `json:"logo,omitempty"`
	Description string                 `json:"description,omitempty"`
	Status      string                 `json:"status"`
	Plan        string                 `json:"plan"`
	MaxUsers    int                    `json:"max_users"`
	MaxStorage  int                    `json:"max_storage_mb"`
	Config      map[string]interface{} `json:"config,omitempty"`
}