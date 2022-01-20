package engine

import (
	"fmt"
	"intery/server/engine/auth/gitea"
	"intery/server/engine/auth/gitee"
	"intery/server/engine/auth/github"
	"intery/server/engine/deploy"
	"intery/server/engine/hi"
	"intery/server/engine/log"
	"intery/server/engine/me"
	"intery/server/engine/project"
	"intery/server/middlewares"

	"github.com/getsentry/sentry-go"
	"github.com/gin-gonic/gin"
)

func NewRouter() *gin.Engine {
	if err := sentry.Init(sentry.ClientOptions{
		Dsn: "https://examplePublicKey@o0.ingest.sentry.io/0",
	}); err != nil {
		fmt.Printf("Sentry initialization failed: %v\n", err)
	}
	router := gin.Default()
	router.Use(middlewares.New()...)
	h := hi.Controller{}
	router.GET("/hi", h.Show)
	api := router.Group("api")
	{
		v1 := api.Group("v1")
		{
			m := me.Controller{}
			v1.GET("/me", m.Show)
			v1.GET("/hi", h.Show)
			deploymentsGroup := v1.Group("deployments")
			{
				d := deploy.Controller{}
				deploymentsGroup.POST("/", d.Create)
				deploymentsGroup.GET("/", d.Index)
			}
			projectsGroup := v1.Group("projects")
			{
				p := project.Controller{}
				projectsGroup.POST("/", p.Create)
				projectsGroup.GET("/", p.Index)
				projectsGroup.GET("/:id", p.Show)
				projectsGroup.DELETE("/:id", p.Delete)
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
			logGroup := v1.Group("logs")
			{
				l := log.Controller{}
				logGroup.GET("", l.Index)
			}
		}
	}
	return router
}
