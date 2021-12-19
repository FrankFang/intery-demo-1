package database

import (
	"fmt"
	"intery/db/query"
	"os"
	"sync"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type container struct {
	mu       sync.Mutex
	instance *gorm.DB
}

var c = container{}

func GetDsnString() string {
	host, user, name, password, port := GetDsn()
	return fmt.Sprintf("host=%s user=%s database=%s password=%s port=%s", host, user, name, password, port)
}
func GetDsn() (host, user, name, password, port string) {
	host = os.Getenv("DB_HOST")
	if host == "" {
		panic("DB_HOST is not set")
	}
	user = os.Getenv("DB_USER")
	if user == "" {
		panic("DB_USER is not set")
	}
	name = os.Getenv("DB_NAME")
	if name == "" {
		panic("DB_NAME is not set")
	}
	password = os.Getenv("DB_PASSWORD")
	if password == "" {
		panic("DB_PASSWORD is not set")
	}
	port = os.Getenv("DB_PORT")
	if port == "" {
		panic("DB_PORT is not set")
	}
	return
}

func Init() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.instance != nil {
		return nil
	}
	dsn := GetDsnString()
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}
	c.instance = db
	return nil
}

func GetDB() *gorm.DB {
	err := Init()
	if err != nil {
		panic(err)
	}
	return c.instance
}
func CloseDB() (err error) {
	db := GetDB()
	sqlconn, err := db.DB()
	if err != nil {
		return
	}
	c.instance = nil
	return sqlconn.Close()
}
func GetQuery() *query.Query {
	return query.Use(GetDB())
}
