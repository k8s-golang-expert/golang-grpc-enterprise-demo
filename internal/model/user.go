package model

import "time"

type User struct {
	ID           int64  `gorm:"primaryKey;autoIncrement"`
	Name         string `gorm:"size:128;not null"`
	Email        string `gorm:"size:256;uniqueIndex;not null"`
	PasswordHash string `gorm:"size:256;not null"`
	CreatedAt    time.Time
}
