package server

import (
	"fmt"
	"io/ioutil"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

func Run() error {
	gin.SetMode("debug")
	app := gin.Default()
	app.SetTrustedProxies(nil)
	app.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			fmt.Println("=====================================origin")
			fmt.Println(origin)
			return origin == "http://localhost:8080"
		},
		MaxAge: 12 * time.Hour,
	}))
	app.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong pong",
		})
	})
	conf := &oauth2.Config{
		ClientID:     "c509b5c3f08700791d87",
		ClientSecret: "cdcd8662ff64410639c068c2eab51e2879060ecb",
		Scopes:       []string{"repo"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://github.com/login/oauth/authorize",
			TokenURL: "https://github.com/login/oauth/access_token",
		},
	}
	app.GET("/demo", func(c *gin.Context) {

		// Redirect user to consent page to ask for permission
		// for the scopes specified above.
		url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
		c.JSON(200, gin.H{
			"message": url,
		})
	})
	app.GET("/auth/github_callback", func(c *gin.Context) {
		code, hasCode := c.GetQuery("code")
		if hasCode {
			tok, err := conf.Exchange(c, code)
			if err != nil {
				fmt.Println(err)
			}

			client := conf.Client(c, tok)
			response, err := client.Get("https://api.github.com/user")
			if err != nil {
				fmt.Println(err)
			}
			bytes, err := ioutil.ReadAll(response.Body)
			if err != nil {
				fmt.Println(err)
			}
			w := c.Writer
			header := w.Header()
			header["Content-Type"] = []string{"application/json;"}
			w.Write(bytes)
		}
	})
	return app.Run()
}
