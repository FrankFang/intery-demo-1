package models

import (
	"intery/server/database"
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`
}

func (BaseModel) TableName() string {
	return "base_models"
}

func (m *BaseModel) Save() error {
	db := database.GetDB()
	if db.Model(&m).Where("id = ?", m.ID).Updates(&m).RowsAffected == 0 {
		return db.Create(&m).Error
	}
	return nil
}

func (m *BaseModel) Create() error {
	return database.GetDB().Create(&m).Error
}
func (m *BaseModel) Update() error {
	return database.GetDB().Save(&m).Error
}
