package github

import (
	"encoding/json"
	"fmt"
	"intery/server/models"
	"io/ioutil"
	"log"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
	"golang.org/x/oauth2"
)

type Controller struct{}
type GitHubUser struct {
	Login     string `json:"login"`
	Id        int64  `json:"id"`
	AvatarUrl string `json:"avatar_url"`
	Name      string `json:"name"`
}

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
	token, err := conf.Exchange(c, code)
	if err != nil {
		fmt.Println(err)
	}
	a := models.Authorization{}
	a.Token = token.AccessToken
	a.RefreshToken = token.RefreshToken
	a.Expiry = token.Expiry
	a.Provider = "github"
	err = a.Save()
	if err != nil {
		log.Fatal(err)
	}

	client := conf.Client(c, token)
	defer client.CloseIdleConnections()
	for i := 0; i < 3; i++ {
		response, err := client.Get("https://api.github.com/user")
		if err != nil {
			continue
		}
		bytes, err := ioutil.ReadAll(response.Body)
		githubUser := GitHubUser{}
		err = json.Unmarshal(bytes, &githubUser)
		if err != nil {
			continue
		}
		a.AvatarUrl = githubUser.AvatarUrl
		a.Name = githubUser.Name
		a.UserId = githubUser.Id
		a.Login = githubUser.Login
		err = a.Save()
		if err != nil {
			panic(err)
			return
		}
		w := c.Writer
		w.Header().Set("Content-Type", "application/json;charset=utf-8")
		_, _ = w.Write(bytes)
		break
	}
}
