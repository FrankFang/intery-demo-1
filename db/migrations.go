package db

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewMigrate(name string) *gormigrate.Gormigrate {
	if gin.Mode() == gin.TestMode {
		name = fmt.Sprintf("%s_test", name)
	} else if gin.Mode() == gin.DebugMode {
		name = fmt.Sprintf("%s_development", name)
	} else {
		name = fmt.Sprintf("%s_production", name)
	}
	dsn := fmt.Sprintf("host=psql1 user=intery database=%v password=123456 port=5432", name)
	// TODO 把所有 Open 合并到一个函数中
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}
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
