package model

import (
	"time"
)

type Authorization struct {
	BaseModel
	Provider         string    `gorm:"type:varchar(100);not null"`
	UserId           uint      `gorm:"not null"`
	VendorId         string    `gorm:"type:varchar(100);not null"`
	Login            string    `gorm:"type:varchar(100);not null"`
	Name             string    `gorm:"type:varchar(100)"`
	AvatarUrl        string    `gorm:"type:text"`
	ReposUrl         string    `gorm:"type:text"`
	Raw              string    `gorm:"type:text"`
	AccessToken      string    `gorm:"type:varchar(100);not null"`
	TokenType        string    `gorm:"type:varchar(100)"`
	RefreshToken     string    `gorm:"type:varchar(100)"`
	Expiry           time.Time `gorm:"default: null"`
	TokenGeneratedAt time.Time `gorm:"not null;default: null"`
}
