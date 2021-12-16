package github

import (
	"encoding/json"
	"fmt"
	"intery/server/database"
	"intery/server/models"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dchest/uniuri"
	"github.com/gin-gonic/gin"
)

type Controller struct{}
type GitHubUser struct {
	Login     string `json:"login"`
	Id        int64  `json:"id"`
	AvatarUrl string `json:"avatar_url"`
	Name      string `json:"name"`
}

var conf = Conf

func (ctrl Controller) Show(c *gin.Context) {
	url := conf.AuthCodeURL(uniuri.New())
	c.JSON(200, gin.H{
		"url": url,
	})
}

func (ctrl Controller) Callback(c *gin.Context) {
	var p struct {
		Code string `json:"code"`
	}
	body, _ := ioutil.ReadAll(c.Request.Body)
	err := json.Unmarshal(body, &p)
	if err != nil {
		c.JSON(400, gin.H{
			"reason": "no code",
		})
		return
	}
	// exchange code for token
	token, err := conf.Exchange(c, p.Code)
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
		auth := models.Authorization{
			Provider: "github",
			Login:    githubUser.Login,
		}
		var user models.User
		database.GetDB().FirstOrInit(&auth)
		if auth.UserId == 0 {
			name := githubUser.Name
			if name == "" {
				name = githubUser.Login
			}
			user = models.User{Name: name}
			if err = user.Create(); err != nil {
				panic(err)
			}
			auth.UserId = user.ID
		} else {
			database.GetDB().First(user, auth.UserId)
		}
		auth.AccessToken = token.AccessToken
		auth.TokenType = token.TokenType
		auth.RefreshToken = token.RefreshToken
		auth.Expiry = token.Expiry
		auth.TokenGeneratedAt = time.Now()
		auth.AvatarUrl = githubUser.AvatarUrl
		auth.Name = githubUser.Name
		auth.VendorId = fmt.Sprintf("%v", githubUser.Id)
		if err = auth.Save(); err != nil {
			panic(err)
		}
		c.JSON(200, gin.H{
			"jwt": user.JWT(),
		})
		break
	}
	if !c.Writer.Written() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"reason": "无法获取 GitHub User 信息",
		})
	}
}
