package db

import (
	"fmt"
	"intery/server/database"
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
					gorm.Model
					Name string `gorm:"type:varchar(100);not null"`
				}
				type Authorization struct {
					gorm.Model
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
				fmt.Println("created table User, Authorzation")
				return d.AutoMigrate(&User{}, &Authorization{})
			},
			Rollback: func(d *gorm.DB) error {
				fmt.Println("dropped table User, Authorzation")
				return d.Migrator().DropTable("users", "authorizations")
			},
		},
		{
			ID: "1639580496128",
			Migrate: func(d *gorm.DB) error {
				type Project struct {
					gorm.Model
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
	})
	return migrations
}
