package models

import (
	"intery/server/database"
)

type Project struct {
	BaseModel `gorm:"embedded"`
	RepoName  string `gorm:"type:varchar(100)" json:"repo_name"`
	AppKind   string `gorm:"type:varchar(100)" json:"app_kind"`
	RepoHome  string `gorm:"type:varchar(1024)" json:"repo_home"`
	UserId    uint   `gorm:"not null" json:"user_id"`
}

func (p *Project) Create() error {
	return database.GetDB().Create(p).Error
}

func (p *Project) Update() error {
	return database.GetDB().Save(p).Error
}
