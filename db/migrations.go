package db

import (
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func NewMigrate() *gormigrate.Gormigrate {
	dsn := "host=psql1 user=intery database=intery_development password=123456 port=5432"
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
					Provider         string   `gorm:"type:varchar(100);not null"`
					UserId           int64    `gorm:"type:bigint"`
					Login            string   `gorm:"type:varchar(100)"`
					Name             string   `gorm:"type:varchar(100)"`
					AvatarUrl        string   `gorm:"type:text"`
					ReposUrl         string   `gorm:"type:text"`
					Raw              struct{} `gorm:"type:jsonb"`
					Token            string   `gorm:"type:varchar(100)"`
					TokenGeneratedAt time.Time
				}
				return d.AutoMigrate(&User{}, &Authorization{})
			},
			Rollback: func(d *gorm.DB) error {
				return d.Migrator().DropTable("users", "authorizations")
			},
		},
	})
	return migrations
}
