package models

import (
	"database/sql"
	"intery/server/database"
	"time"
)

type Project struct {
	ID        uint         `gorm:"primarykey" json:"id"`
	CreatedAt time.Time    `json:"created_at"`
	UpdatedAt time.Time    `json:"updated_at"`
	DeletedAt sql.NullTime `gorm:"index" json:"deleted_at"`
	RepoName  string       `gorm:"type:varchar(100)" json:"repo_name"`
	AppKind   string       `gorm:"type:varchar(100)" json:"app_kind"`
	RepoHome  string       `gorm:"type:varchar(1024)" json:"repo_home"`
	UserId    uint         `gorm:"not null" json:"user_id"`
}

func (p *Project) Create() error {
	return database.GetDB().Create(p).Error
}

func (p *Project) Update() error {
	return database.GetDB().Save(p).Error
}
