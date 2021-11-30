package server

import "github.com/gin-gonic/gin"

func Run() error {
	app := gin.Default()
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	return app.Run()
}
