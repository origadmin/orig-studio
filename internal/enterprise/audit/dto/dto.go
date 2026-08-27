package dto

import "time"

type Entry struct {
	ID         string    `json:"id,omitempty"`
	UserID     string    `json:"user_id,omitempty"`
	Username   string    `json:"username,omitempty"`
	Action     string    `json:"action,omitempty"`
	Resource   string    `json:"resource,omitempty"`
	IP         string    `json:"ip,omitempty"`
	UserAgent  string    `json:"user_agent,omitempty"`
	Result     string    `json:"result,omitempty"`
	Detail     string    `json:"detail,omitempty"`
	CreateTime time.Time `json:"create_time,omitempty"`
}

type QueryFilter struct {
	UserID    string `json:"user_id,omitempty"`
	Action    string `json:"action,omitempty"`
	Resource  string `json:"resource,omitempty"`
	Result    string `json:"result,omitempty"`
	StartTime string `json:"start_time,omitempty"`
	EndTime   string `json:"end_time,omitempty"`
	Page      int    `json:"page,omitempty"`
	PageSize  int    `json:"page_size,omitempty"`
}

type QueryResult struct {
	Items    []*Entry `json:"items"`
	Total    int      `json:"total"`
	Page     int      `json:"page"`
	PageSize int      `json:"page_size"`
}