package engine

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func Init(db *gorm.DB) (*gin.Engine, error) {
	gin.SetMode("debug")
	r := NewRouter()
	r.Use(func(c *gin.Context) {
		c.Set("db", db)
		
	})
	err := r.SetTrustedProxies(nil)
	if err != nil {
		return nil, err
	}

	return r, nil
}
