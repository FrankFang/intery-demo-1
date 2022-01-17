package base

import (
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"intery/server/database"
	"intery/server/engine/auth/github"
	"intery/server/model"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	sdk "github.com/google/go-github/v41/github"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"golang.org/x/oauth2"
)

type BaseController struct {
	// userId uint
	user *model.User
	auth *model.Authorization
}

func (ctrl *BaseController) MustHasPage(c *gin.Context) (page, perPage, offset int, err error) {
	offsetString := c.DefaultQuery("offset", "0")
	pageString := c.DefaultQuery("page", "1")
	perPageString := c.DefaultQuery("per_page", "10")
	page, err = strconv.Atoi(pageString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"reason": "page 必须是数字",
		})
		return
	}
	perPage, err = strconv.Atoi(perPageString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"reason": "per_page 必须是数字",
		})
		return
	}
	offset, err = strconv.Atoi(offsetString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"reason": "offset 必须是数字",
		})
		return
	}
	return
}

func (ctrl *BaseController) MustSignIn(c *gin.Context) (userId uint, err error) {
	userId, err = ctrl.GetUserIdFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"reason": err.Error(),
		})
		return
	}
	return
}

func (ctrl *BaseController) GetUserIdFromHeader(c *gin.Context) (uint, error) {
	// FIXME 不知道为什么 ctrl.userId 不会在请求结束后被清空
	// if ctrl.userId != 0 {
	// 	return ctrl.userId, nil
	// }
	bearer := c.Request.Header.Get("Authorization")
	parts := strings.Split(bearer, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		err := errors.New("invalid authorization")
		return 0, err
	}
	jwtString := parts[1]
	token, err := jwt.Parse(jwtString, keyFunc)
	if err != nil {
		err = errors.New("invalid jwt")
		return 0, err
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("invalid claims")
		return 0, err
	}
	// ctrl.userId = uint(claims["user_id"].(float64))
	// return ctrl.userId, nil
	return uint(claims["user_id"].(float64)), nil
}

func (ctrl *BaseController) MustAuth(c *gin.Context) (auth *model.Authorization, err error) {
	userId, err := ctrl.MustSignIn(c)
	if err != nil {
		return
	}
	auth, err = ctrl.GetAuthFromUserId(userId)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"reason": err.Error(),
		})
	}
	return
}

func (ctrl *BaseController) GetAuthFromUserId(userId uint) (auth *model.Authorization, err error) {
	if ctrl.auth != nil {
		return ctrl.auth, nil
	}
	err = database.GetDB().Find(&ctrl.auth, "user_id = ?", userId).Error
	return ctrl.auth, err
}

func (ctrl *BaseController) GetUserIdAndAuth(c *gin.Context) (userId uint, auth *model.Authorization, err error) {
	userId, err = ctrl.GetUserIdFromHeader(c)
	if err != nil {
		err = fmt.Errorf("无法从 Header 中获取 user id")
		return
	}
	auth, err = ctrl.GetAuthFromUserId(userId)
	if err != nil {
		return
	}
	return
}

func (ctrl *BaseController) GetUserAndAuth(c *gin.Context) (user *model.User, auth *model.Authorization, err error) {
	userId, auth, err := ctrl.GetUserIdAndAuth(c)
	if err != nil {
		err = fmt.Errorf("无法从 Header 中获取 user id")
		return
	}
	if ctrl.user != nil {
		user = ctrl.user
	} else {
		u := database.GetQuery().User
		user, err = u.WithContext(c).Where(u.ID.Eq(userId)).First()
		if err != nil {
			err = fmt.Errorf("不存在 id 为 %v 的用户", userId)
			return
		}
	}
	return
}
func (ctrl *BaseController) GetGithubClient(c *gin.Context, auth *model.Authorization) (client *sdk.Client) {
	oauth2Token := oauth2.Token{AccessToken: auth.AccessToken, RefreshToken: "-"}
	client = sdk.NewClient(github.Conf.Client(c, &oauth2Token))
	return
}
func (ctrl *BaseController) HandleGitHubError(c *gin.Context, err error) {
	if resErr, ok := err.(*sdk.ErrorResponse); ok {
		defer resErr.Response.Body.Close()
		transformResponse(c, resErr.Response)
	} else if err.Error() == "unexpected status code: 404 Not Found" {
		c.AbortWithStatus(http.StatusNotFound)
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
	}
}

// helper
func keyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
		log.Println("method not supported")
	}
	key, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		log.Println(err)
	}
	block, _ := pem.Decode(key)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Println("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Println(err)
	}
	return pub, nil
}

func transformResponse(c *gin.Context, response *http.Response) {
	if response.ContentLength == 0 {
		c.JSON(http.StatusInternalServerError,
			gin.H{"reason": "GitHub API error status code " + response.Status})
		return
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Println("Read response body failed.", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	var data struct {
		DocumentationUrl string `json:"documentation_url"`
		Errors           []struct {
			Code     string
			Field    string
			Message  string
			Resource string
		}
		Message string `json:"message"`
	}
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Println("Unmarshal response body failed.", content, err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	if data.Message == "Repository creation failed." {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{
			"repo_name": []string{"GitHub上已经存在该仓库，请使用其他仓库名。"},
		}})
		return
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": data.Message})
	}
}
