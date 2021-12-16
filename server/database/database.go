package database

import (
	"intery/db/query"
	"sync"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type container struct {
	mu       sync.Mutex
	instance *gorm.DB
}

var c = container{}

func Init() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.instance != nil {
		return nil
	}
	dsn := "host=psql1 user=intery database=intery_development password=123456 port=5432"
	if gin.Mode() == gin.TestMode {
		dsn = "host=psql1 user=intery database=intery_test password=123456 port=5432"
	}
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
func GetQuery() *query.Query {
	return query.Use(GetDB())
}
