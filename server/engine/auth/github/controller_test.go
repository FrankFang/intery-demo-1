package github

import (
	"encoding/json"
	"intery/server/test"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/h2non/gock.v1"
)

type GitHubControllerTestSuite struct {
	suite.Suite
}

func TestGitHubCtrollerTestSuite(t *testing.T) {
	test.Setup(t)
	suite.Run(t, new(GitHubControllerTestSuite))
	test.Teardown(t)
}

func (s *GitHubControllerTestSuite) TestShow() {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	ctrl := Controller{}
	ctrl.Show(c)
	assert.Equal(s.T(), 200, w.Code)

	var body gin.H
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.IsType(s.T(), "", body["url"])
}

func (s *GitHubControllerTestSuite) TestCallback() {
	gock.New("https://github.com").
		Post("/login/oauth/access_token").
		Reply(200).
		JSON(map[string]interface{}{"access_token": "token"})
	gock.New("https://api.github.com").
		Get("/user").
		Reply(200).
		JSON(map[string]interface{}{"login": "test", "id": 1, "avatar_url": "test", "name": "高圆圆"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/auth/github_callback", strings.NewReader(`{"code": "123", "state": "123"}`))
	ctrl := Controller{}
	ctrl.Callback(c)
	assert.Equal(s.T(), 200, w.Code)

	var body gin.H
	err := json.Unmarshal(w.Body.Bytes(), &body)
	if err != nil {
		s.T().Fatal(err)
	}
	assert.IsType(s.T(), "", body["jwt"])
}
