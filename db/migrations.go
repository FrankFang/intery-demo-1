package db

import (
	"fmt"
	"intery/server/database"
	"intery/server/model"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

func NewMigrate() *gormigrate.Gormigrate {
	db := database.GetDB()
	migrations := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "1638802075376",
			Migrate: func(d *gorm.DB) error {
				type User struct {
					model.BaseModel
					Name string `gorm:"type:varchar(100);not null"`
				}
				type Authorization struct {
					model.BaseModel
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
				fmt.Println("created table User")
				fmt.Println("created table Authorzation")
				return d.AutoMigrate(&User{}, &Authorization{})
			},
			Rollback: func(d *gorm.DB) error {
				fmt.Println("dropped table User")
				fmt.Println("dropped table Authorzation")
				return d.Migrator().DropTable("users", "authorizations")
			},
		},
		{
			ID: "1639580496128",
			Migrate: func(d *gorm.DB) error {
				type Project struct {
					model.BaseModel
					RepoName string `gorm:"type:varchar(100)"`
					AppKind  string `gorm:"type:varchar(100)"`
					RepoHome string `gorm:"type:varchar(1024)"`
					UserId   uint   `gorm:"not null"`
				}
				fmt.Println("created table Project")
				return d.AutoMigrate(&Project{})
			},
			Rollback: func(d *gorm.DB) error {
				fmt.Println("dropped table Project")
				return d.Migrator().DropTable("projects")
			},
		},
		{
			ID: "1640795201612",
			Migrate: func(d *gorm.DB) error {
				type Deployment struct {
					model.BaseModel
					ContainerId uint `gorm:"not null"`
					ProjectId   uint `gorm:"not null"`
				}
				fmt.Println("created table Deployment")
				return d.AutoMigrate(&Deployment{})
			},
			Rollback: func(d *gorm.DB) error {
				fmt.Println("dropped table Deployment")
				return d.Migrator().DropTable("deployments")
			},
		},
	})
	return migrations
}
