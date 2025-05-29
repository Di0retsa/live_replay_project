package model

import "time"

type ChatMessage struct {
	MessageId uint32    `json:"message_id,omitempty"`
	UserId    uint32    `json:"user_id,omitempty"`
	Username  string    `json:"username,omitempty"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	ReplayId  uint32    `json:"replay_id,omitempty"`
	Type      string    `json:"type,omitempty"`
}
