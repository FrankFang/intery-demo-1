package project

import (
	"encoding/json"
	"fmt"
	"intery/server/database"
	"intery/test"
	"intery/test/helper"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type ProjectControllerTestSuite struct {
	suite.Suite
}

func TestProjectControllerTestSuite(t *testing.T) {
	id := test.GetId()
	test.Setup(t, id)
	suite.Run(t, new(ProjectControllerTestSuite))
	test.Teardown(t, id)
}

func (s *ProjectControllerTestSuite) TestCreate() {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	user := helper.CreateUser(c)
	gock.New("https://api.github.com").
		Post("/user/repos").
		Reply(200).
		JSON(map[string]interface{}{"access_token": "token"})
	gock.New("https://api.github.com").
		Put(fmt.Sprintf("repos/%s/%s/contents/%s", user.Name, "test", "README.md")).
		Reply(200).JSON(map[string]interface{}{"commit": map[string]interface{}{"sha": "sha"}})
	gock.New("https://api.github.com").
		Get(fmt.Sprintf("repos/%v/%v/git/trees/%v", user.Name, "test", "sha")).
		Reply(200).JSON(map[string]interface{}{"sha": "sha"})
	gock.New("https://api.github.com").
		Post(fmt.Sprintf("repos/%v/%v/git/trees", user.Name, "test")).
		Reply(200).JSON(map[string]interface{}{"sha": "sha"})
	gock.New("https://api.github.com").
		Post(fmt.Sprintf("repos/%v/%v/git/commits", user.Name, "test")).
		Reply(200).JSON(map[string]interface{}{"sha": "sha"})
	gock.New("https://api.github.com").
		Patch(fmt.Sprintf("repos/%v/%v/git/refs/%v", user.Name, "test", "refs/heads/main")).
		Reply(200).
		JSON(map[string]interface{}{"ref": "refs/heads/main"})
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/github_callback",
		strings.NewReader(`{"app_kind": "nodejs", "repo_name": "test"}`))
	helper.SignIn(c, user)
	count1, _ := database.GetQuery().Project.WithContext(c).Count()
	ctrl := Controller{}
	ctrl.Create(c)
	assert.Equal(s.T(), http.StatusCreated, w.Code)

	var body gin.H
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		s.T().Fatal(err)
	}
	resource := body["resource"].(map[string]interface{})
	assert.IsType(s.T(), 1.23, resource["id"])
	count2, _ := database.GetQuery().Project.WithContext(c).Count()
	// 数据库中新增一个 project
	assert.Equal(s.T(), count1+1, count2)
}
