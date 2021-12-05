package engine

import (
	"intery/server/engine/auth/github"
	"intery/server/engine/hi"
	"intery/server/middlewares"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.New()...)
	h := hi.Controller{}
	router.GET("/hi", h.Show)
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
