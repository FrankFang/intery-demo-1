package gitea

import (
	"fmt"
	"io/ioutil"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Controller struct{}

var conf = &oauth2.Config{
	ClientID:     "5ce02d15-2b31-41e6-8aa8-3f59a9e7b597",
	ClientSecret: "VR1w1qyUP21clz4x9qPQjwSQ4iYRL9rpTDpeFP1SCO7K",
	RedirectURL:  "http://intery.xiedaimala.com/api/v1/auth/gitea_callback",
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://gitea.com/login/oauth/authorize",
		TokenURL: "https://gitea.com/login/oauth/access_token",
	},
}

func (ctrl *Controller) Show(c *gin.Context) {
	url := conf.AuthCodeURL(uniuri.New())
	c.JSON(200, gin.H{
		"url": url,
	})
}

func (ctrl Controller) Callback(c *gin.Context) {
	code, hasCode := c.GetQuery("code")
	if !hasCode {
		c.JSON(400, gin.H{
			"reason": "no code",
		})
		return
	}
	tok, err := conf.Exchange(c, code)
	if err != nil {
		fmt.Println(err)
	}

	client := conf.Client(c, tok)
	response, err := client.Get("https://gitea.com/api/v1/user")
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
