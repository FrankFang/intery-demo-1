package base

import (
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"intery/server/database"
	"intery/server/engine/auth/github"
	"intery/server/model"
	"io/ioutil"
	"log"
	"os"
	"strings"

	sdk "github.com/google/go-github/v41/github"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
)

type BaseController struct{}

func (ctrl *BaseController) GetUserIdFromHeader(c *gin.Context) (userId uint, err error) {
	bearer := c.Request.Header.Get("Authorization")
	parts := strings.Split(bearer, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		err = errors.New("invalid authorization")
		return
	}
	jwtString := parts[1]
	token, err := jwt.Parse(jwtString, keyFunc)
	if err != nil {
		err = errors.New("invalid jwt")
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("invalid claims")
		return
	}
	userId = uint(claims["user_id"].(float64))
	return
}

func (ctrl *BaseController) GetAuthFromUserId(userId uint) (auth *model.Authorization) {
	database.GetDB().Find(&auth, "user_id = ?", userId)
	return auth
}

func (ctrl *BaseController) GetUserAndAuth(c *gin.Context) (user *model.User, auth *model.Authorization, err error) {
	userId, err := ctrl.GetUserIdFromHeader(c)
	if err != nil {
		err = fmt.Errorf("无法从 Header 中获取 user id")
		return
	}
	u := database.GetQuery().User
	user, err = u.WithContext(c).Where(u.ID.Eq(userId)).First()
	if err != nil {
		err = fmt.Errorf("不存在 id 为 %v 的用户", userId)
		return
	}
	auth = ctrl.GetAuthFromUserId(userId)
	if auth == nil {
		err = errors.New("未注册")
		return
	}
	return
}
func (ctrl *BaseController) GetGithubClient(c *gin.Context, auth *model.Authorization) (client *sdk.Client) {
	oauth2Token := oauth2.Token{AccessToken: auth.AccessToken, RefreshToken: "-"}
	client = sdk.NewClient(github.Conf.Client(c, &oauth2Token))
	return
}

// helper
func keyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
		log.Fatal("method not supported")
	}
	key, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	block, _ := pem.Decode(key)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	return pub, nil
}
