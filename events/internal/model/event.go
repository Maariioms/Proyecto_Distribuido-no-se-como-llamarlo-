package model

import "time"

type Event struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Type      string    `gorm:"not null" json:"type"`
	Severity  string    `gorm:"not null" json:"severity"`
	CreatedAt time.Time `json:"created_at"`
}
