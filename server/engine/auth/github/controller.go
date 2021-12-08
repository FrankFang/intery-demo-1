package github

import (
	"encoding/json"
	"fmt"
	"intery/server/models"
	"io/ioutil"
	"time"

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
	// get code
	code, hasCode := c.GetQuery("code")
	if !hasCode {
		c.JSON(400, gin.H{
			"reason": "no code",
		})
		return
	}
	// exchange code for token
	token, err := conf.Exchange(c, code)
	if err != nil {
		fmt.Println(err)
	}
	// create client with token
	client := conf.Client(c, token)
	defer client.CloseIdleConnections()

	// get github user via client
	for i := 0; i < 3; i++ {
		response, err := client.Get("https://api.github.com/user")
		if err != nil {
			continue
		}
		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			continue
		}
		githubUser := GitHubUser{}
		err = json.Unmarshal(bytes, &githubUser)
		if err != nil {
			continue
		}
		// create authorization
		name := githubUser.Name
		if name == "" {
			name = githubUser.Login
		}
		user := models.User{Name: name}
		err = user.Create()
		if err != nil {
			panic(err)
		}
		a := models.Authorization{
			UserId:           user.ID,
			Token:            token.AccessToken,
			TokenGeneratedAt: time.Now(),
			TokenType:        token.TokenType,
			RefreshToken:     token.RefreshToken,
			Expiry:           token.Expiry,
			Provider:         "github",
			AvatarUrl:        githubUser.AvatarUrl,
			Name:             githubUser.Name,
			VendorId:         fmt.Sprintf("%v", githubUser.Id),
			Login:            githubUser.Login,
		}
		err = a.Save()
		if err != nil {
			panic(err)
		}
		// c.JSON(200, gin.H{
		// 	"jwt": user.JWT(),
		// })
		c.HTML(200, "jwt.html", gin.H{
			"jwt":      user.JWT(),
			"returnTo": "http://127.0.0.1:3000",
		})
		break
	}
}
