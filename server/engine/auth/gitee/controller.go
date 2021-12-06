package gitee

import (
	"fmt"
	"io/ioutil"

	"github.com/dchest/uniuri"

	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Controller struct{}

var conf = &oauth2.Config{
	ClientID:     "ee73eb425c78c965e4b749ec7fb250beee30449fac584ac8b13e0e2a92dd74fe",
	ClientSecret: "61484b0b39a3d974dbd031b149e152fd469f27368ae600ec804eb438d4c4fafe",
	Scopes:       []string{"projects", "emails"},
	RedirectURL:  "http://intery.xiedaimala.com/api/v1/auth/gitee_callback",
	Endpoint: oauth2.Endpoint{
		AuthURL:  "https://gitee.com/oauth/authorize",
		TokenURL: "https://gitee.com/oauth/token",
	},
}

func (ctrl Controller) Show(c *gin.Context) {
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
	response, err := client.Get("https://gitee.com/api/v5/user")
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
