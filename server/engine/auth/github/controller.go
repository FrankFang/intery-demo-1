package github

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
	"io/ioutil"
)

type Controller struct{}

var conf = &oauth2.Config{
	ClientID:     "c509b5c3f08700791d87",
	ClientSecret: "cdcd8662ff64410639c068c2eab51e2879060ecb",
	Scopes:       []string{"repo"},
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://github.com/login/oauth/authorize",
		TokenURL: "https://github.com/login/oauth/access_token",
	},
}

func (ctrl Controller) Show(c *gin.Context) {
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	c.JSON(200, gin.H{
		"message": url,
	})
}

func (ctrl Controller) Callback(c *gin.Context) {
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
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		_, _ = w.Write(bytes)
	}
}
