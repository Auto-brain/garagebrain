package model

import "time"

type UserIdentity struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	Platform    string    `json:"platform"`
	PlatformID  string    `json:"platform_id"`
	Username    string    `json:"username,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	LinkedAt    time.Time `json:"linked_at"`
}

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Platform  string    `json:"platform"`
	ChatID    string    `json:"chat_id"`
	Service   string    `json:"service"`
	Messages  []byte    `json:"messages,omitempty"`
	Profile   []byte    `json:"profile,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}
