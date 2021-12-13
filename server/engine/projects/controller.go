package projects

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"intery/server/database"
	"intery/server/engine/auth/github"
	"intery/server/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	sdk "github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type Controller struct {
}

func transformResponse(c *gin.Context, response *http.Response) {
	c.Status(response.StatusCode)
	if response.ContentLength == 0 {
		return
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Read response body failed.", err)
	}
	c.Writer.Write(content)
}

func (ctrl *Controller) Create(c *gin.Context) {
	var p struct {
		AppKind  string `json:"app_kind" binding:"required"`
		RepoName string `json:"repo_name" binding:"required"`
	}
	if err := c.BindJSON(&p); err != nil {
		log.Fatal(err)
	}
	bearer := c.Request.Header.Get("Authorization")
	parts := strings.Split(bearer, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		c.JSON(401, gin.H{"reason": "invalid authorization"})
		return
	}
	jwtString := parts[1]
	token, err := jwt.Parse(jwtString, keyFunc)
	if err != nil {
		c.JSON(401, gin.H{"reason": "invalid jwt"})
		return
	}
	fmt.Printf("%+v \n", token.Claims)
	claims := token.Claims.(jwt.MapClaims)
	auth := models.Authorization{}
	database.GetDB().Find(&auth, "user_id = ?", int(claims["user_id"].(float64)))
	oToken := oauth2.Token{AccessToken: auth.AccessToken, RefreshToken: "hi"}
	client := github.Conf.Client(c, &oToken)
	client2 := sdk.NewClient(client)
	repo, response, err := client2.Repositories.Create(c, "", &sdk.Repository{
		Name:    sdk.String(p.RepoName),
		Private: sdk.Bool(false),
	})
	if err != nil {
		if err, ok := err.(*sdk.ErrorResponse); ok {
			defer err.Response.Body.Close()
			transformResponse(c, err.Response)
			return
		} else {
			log.Fatal("Create repo failed.", err)
		}
	}
	ref, response2, err := client2.Git.GetRef(c, auth.Login, repo.GetName(), "HEAD")
	fmt.Println(ref, response2, err)
	// repo is empty
	c.JSON(response.StatusCode, repo)
}

// helper function
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
