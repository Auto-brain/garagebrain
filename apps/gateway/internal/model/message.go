package model

import "time"

type IncomingMessage struct {
	Platform    string    `json:"platform"`
	UserID      string    `json:"user_id"`
	ChatID      string    `json:"chat_id"`
	Username    string    `json:"username,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	Text        string    `json:"text"`
	Photos      []string  `json:"photos,omitempty"`
	ReplyTo     string    `json:"reply_to,omitempty"`
	ReceivedAt  time.Time `json:"received_at"`
}

type OutgoingMessage struct {
	ChatID    string     `json:"chat_id"`
	Text      string     `json:"text"`
	Buttons   [][]Button `json:"buttons,omitempty"`
	Photo     string     `json:"photo,omitempty"`
	ParseMode string     `json:"parse_mode,omitempty"`
}

type Button struct {
	Text    string `json:"text"`
	Payload string `json:"payload"`
}

type UserIdentity struct {
	ID          int       `json:"id"`
	UserID      string    `json:"user_id"`
	Platform    string    `json:"platform"`
	PlatformID  string    `json:"platform_id"`
	Username    string    `json:"username,omitempty"`
	DisplayName string    `json:"display_name,omitempty"`
	LinkedAt    time.Time `json:"linked_at"`
}
