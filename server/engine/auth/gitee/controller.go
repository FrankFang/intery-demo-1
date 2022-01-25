package gitee

import (
	"encoding/json"
	"fmt"
	"intery/server/config"
	"intery/server/database"
	"intery/server/model"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/dchest/uniuri"

	"github.com/gin-gonic/gin"
)

type Controller struct{}
type GiteeUser struct {
	Login     string `json:"login"`
	Id        int64  `json:"id"`
	AvatarUrl string `json:"avatar_url"`
	Name      string `json:"name"`
}

var conf = Conf

func (ctrl Controller) Show(c *gin.Context) {
	state := uniuri.New()
	config.AddOAuth2State(state)
	url := conf.AuthCodeURL(state)
	c.JSON(200, gin.H{
		"url": url,
	})
}

func (ctrl Controller) Callback(c *gin.Context) {
	var params struct {
		Code  string `json:"code"`
		State string `json:"state"`
	}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}
	if !config.UseOAuth2State(params.State) {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "state 错误"})
		return
	}

	token, err := conf.Exchange(c, params.Code)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"reason": "exchange token via code failed",
		})
		return
	}
	client := conf.Client(c, token)
	defer client.CloseIdleConnections()
	for i := 0; i < 2; i++ {
		response, err := client.Get("https://gitee.com/api/v5/user")
		if err != nil {
			continue
		}
		bytes, err := ioutil.ReadAll(response.Body)
		if err != nil {
			continue
		}
		giteeUser := GiteeUser{}
		err = json.Unmarshal(bytes, &giteeUser)
		if err != nil {
			continue
		}
		u := database.GetQuery().User
		var user *model.User
		a := database.GetQuery().Authorization
		auth, err := a.WithContext(c).Where(a.Login.Eq(giteeUser.Login)).Where(a.Provider.Eq("gitee")).FirstOrInit()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"reason": "Authorization not found. " + err.Error(),
			})
			return
		}
		if auth.UserId == 0 {
			name := giteeUser.Name
			if name == "" {
				name = giteeUser.Login
			}
			user = &model.User{Name: name}
			if err := u.WithContext(c).Create(user); err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"reason": "create user failed. " + err.Error(),
				})
				return
			}
			auth.UserId = user.ID
		} else {
			user, err = u.WithContext(c).Where(u.ID.Eq(auth.UserId)).First()
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"reason": "find user failed. " + err.Error(),
				})
				return
			}
		}
		auth.AccessToken = token.AccessToken
		auth.TokenType = token.TokenType
		auth.RefreshToken = token.RefreshToken
		auth.Expiry = token.Expiry
		auth.TokenGeneratedAt = time.Now()
		auth.AvatarUrl = giteeUser.AvatarUrl
		auth.Name = giteeUser.Name
		auth.VendorId = fmt.Sprintf("%v", giteeUser.Id)
		if err = database.GetQuery().WithContext(c).Authorization.Save(auth); err != nil {
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
			"reason": "无法获取 Gitee User 信息",
		})
	}
}
