package github

import (
	"encoding/json"
	"fmt"
	"intery/server/database"
	"intery/server/model"
	"io/ioutil"
	"log"
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
	body, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		log.Fatal(err)
	}
	err = json.Unmarshal(body, &p)
	if err != nil {
		c.JSON(400, gin.H{
			"reason": "no code",
		})
		return
	}
	token, err := conf.Exchange(c, p.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"reason": "exchange token via code failed",
		})
		return
	}
	client := conf.Client(c, token)
	defer client.CloseIdleConnections()
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
		auth := model.Authorization{
			Provider: "github",
			Login:    githubUser.Login,
		}
		var user model.User
		database.GetDB().FirstOrInit(&auth)
		if auth.UserId == 0 {
			name := githubUser.Name
			if name == "" {
				name = githubUser.Login
			}
			user = model.User{Name: name}
			err := database.GetQuery().WithContext(c).User.Create(&user)
			if err != nil {
				panic(err)
			}
			auth.UserId = user.ID
		} else {
			database.GetDB().First(&user, auth.UserId)
		}
		auth.AccessToken = token.AccessToken
		auth.TokenType = token.TokenType
		auth.RefreshToken = token.RefreshToken
		auth.Expiry = token.Expiry
		auth.TokenGeneratedAt = time.Now()
		auth.AvatarUrl = githubUser.AvatarUrl
		auth.Name = githubUser.Name
		auth.VendorId = fmt.Sprintf("%v", githubUser.Id)
		database.GetQuery().WithContext(c).Authorization.Save(&auth)
		if err = database.GetQuery().WithContext(c).Authorization.Save(&auth); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		}
		if t, err := user.JWT(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		} else {
			c.JSON(http.StatusOK, gin.H{"jwt": t})
		}
		break
	}
	if !c.Writer.Written() {
		c.JSON(http.StatusUnauthorized, gin.H{
			"reason": "无法获取 GitHub User 信息",
		})
	}
}
