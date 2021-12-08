package models

import (
	"intery/server/database"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name string `gorm:"type:varchar(100);not null"`
}

func (u *User) Create() error {
	return database.GetDB().Create(&u).Error
}
func (u *User) Update() error {
	return database.GetDB().Save(&u).Error
}
