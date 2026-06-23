package schema

import "time"

type AuditLog struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Username   string    `json:"username"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	Detail     string    `json:"detail"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	Result     string    `json:"result"`
	CreatedAt  time.Time `json:"created_at"`
}

const (
	ActionLogin    = "login"
	ActionLogout   = "logout"
	ActionCreate   = "create"
	ActionUpdate   = "update"
	ActionDelete   = "delete"
	ActionPayment  = "payment"
	ActionPermChange = "permission_change"

	ResultSuccess = "success"
	ResultFailure = "failure"

	ResourceUser       = "user"
	ResourceMedia      = "media"
	ResourcePayment    = "payment"
	ResourcePermission = "permission"
	ResourceSystem     = "system"
)

const AuditLogTableName = "audit_logs"