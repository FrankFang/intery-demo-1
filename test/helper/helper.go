package helper

import (
	"encoding/json"
	"intery/server/database"
	"intery/server/engine/auth/github"
	"intery/server/model"
	"log"
	"net/http/httptest"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/h2non/gock.v1"
)

func CreateUser(c *gin.Context) (user *model.User) {
	user = &model.User{
		Name: "test",
	}
	err := database.GetQuery().WithContext(c).User.Create(user)
	if err != nil {
		panic(err)
	}
	auth := model.Authorization{
		Provider:         "github",
		Login:            user.Name,
		UserId:           user.ID,
		AccessToken:      "token",
		TokenGeneratedAt: time.Now(),
		VendorId:         "123",
	}
	err = database.GetQuery().WithContext(c).Authorization.Create(&auth)
	if err != nil {
		panic(err)
	}
	return
}

func SignIn(c *gin.Context, user *model.User) {
	if user == nil {
		return
	}
	gock.New("https://github.com").
		Post("/login/oauth/access_token").
		Reply(200).
		JSON(map[string]interface{}{"access_token": "token"})
	gock.New("https://api.github.com").
		Get("/user").
		Reply(200).
		JSON(map[string]interface{}{"login": user.Name, "id": 404, "avatar_url": "test", "name": user.Name})
	w2 := httptest.NewRecorder()
	c2, _ := gin.CreateTestContext(w2)
	c2.Request = httptest.NewRequest("POST", "/api/v1/auth/github_callback",
		strings.NewReader(`{"code": "123", "state": "123"}`))
	ctrl := github.Controller{}
	ctrl.Callback(c2)
	var body gin.H
	json.Unmarshal(w2.Body.Bytes(), &body)
	jwt, ok := body["jwt"].(string)
	if !ok {
		log.Fatal("jwt not found")
	}
	c.Request.Header.Add("Authorization", "Bearer "+jwt)
}
