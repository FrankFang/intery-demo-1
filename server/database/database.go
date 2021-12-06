package database

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var globalDb *gorm.DB

func Init() *gorm.DB {
	dsn := "host=psql1 user=intery database=intery_development password=123456 port=5432"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	globalDb = db
	if err != nil {
		panic("failed to connect database")
	}
	return db
}

func GetDB() *gorm.DB {
	return globalDb
}
