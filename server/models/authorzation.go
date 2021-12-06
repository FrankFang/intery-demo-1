package models

import (
	"intery/server/database"
	"time"

	"gorm.io/gorm"
)

type Authorization struct {
	gorm.Model
	Provider         string   `gorm:"type:varchar(100);not null;default:null"`
	UserId           int64    `gorm:"type:bigint"`
	Login            string   `gorm:"type:varchar(100)"`
	Name             string   `gorm:"type:varchar(100)"`
	AvatarUrl        string   `gorm:"type:text"`
	ReposUrl         string   `gorm:"type:text"`
	Raw              struct{} `gorm:"type:jsonb"`
	Token            string   `gorm:"type:varchar(100)"`
	TokenGeneratedAt time.Time
}

func (a Authorization) Save() error {
	result := database.GetDB().Create(&a)
	return result.Error
}
