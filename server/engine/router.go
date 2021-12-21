package engine

import (
	"intery/server/engine/auth/gitea"
	"intery/server/engine/auth/gitee"
	"intery/server/engine/auth/github"
	"intery/server/engine/deploy"
	"intery/server/engine/hi"
	"intery/server/engine/project"
	"intery/server/middlewares"

	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	router := gin.Default()
	router.Use(middlewares.New()...)
	h := hi.Controller{}
	router.GET("/hi", h.Show)
	api := router.Group("api")
	{
		v1 := api.Group("v1")
		{
			deploysGroups := v1.Group("deploys")
			{
				d := deploy.Controller{}
				deploysGroups.POST("/", d.Create)
			}
			projectsGroup := v1.Group("projects")
			{
				p := project.Controller{}
				projectsGroup.POST("/", p.Create)
			}
			authGroup := v1.Group("auth")
			{
				g := github.Controller{}
				authGroup.GET("/github", g.Show)
				authGroup.POST("/github_callback", g.Callback)
				t := gitee.Controller{}
				authGroup.GET("/gitee", t.Show)
				authGroup.GET("/gitee_callback", t.Callback)
				a := gitea.Controller{}
				authGroup.GET("/gitea", a.Show)
				authGroup.GET("/gitea_callback", a.Callback)
			}
		}
	}
	return router
}
