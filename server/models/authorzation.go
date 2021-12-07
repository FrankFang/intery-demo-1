package models

import (
	"intery/server/database"
	"time"

	"gorm.io/gorm"
)

type Authorization struct {
	gorm.Model
	Provider         string    `gorm:"type:varchar(100);not null;default:null"`
	UserId           int64     `gorm:"type:bigint"`
	Login            string    `gorm:"type:varchar(100)"`
	Name             string    `gorm:"type:varchar(100)"`
	AvatarUrl        string    `gorm:"type:text"`
	ReposUrl         string    `gorm:"type:text"`
	Raw              struct{}  `gorm:"type:jsonb"`
	Token            string    `gorm:"type:varchar(100)"`
	TokenType        string    `gorm:"type:varchar(100)"`
	RefreshToken     string    `gorm:"type:varchar(100)"`
	Expiry           time.Time `gorm:"default: null"`
	TokenGeneratedAt time.Time `gorm:"not null;default: null"`
}

func (a Authorization) Save() error {
	if a.ID == 0 {
		return database.GetDB().Create(&a).Error
	} else {
		return database.GetDB().Save(&a).Error
	}
}
