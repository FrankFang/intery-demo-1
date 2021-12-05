package engine

import (
	"github.com/gin-gonic/gin"
	"intery/server/engine/auth/github"
	"intery/server/engine/hi"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	hiController := hi.Controller{}
	router.GET("/hi", hiController.Show)
	v1 := router.Group("v1")
	{
		authGroup := v1.Group("auth")
		{
			g := github.Controller{}
			authGroup.GET("/github", g.Show)
			authGroup.GET("/github_callback", g.Callback)
		}
	}
	return router
}
