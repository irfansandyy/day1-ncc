package models

import "time"

type Message struct {
	ID        int64     `json:"id"`
	ChatID    int64     `json:"chat_id"`
	Role      string    `json:"role"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
