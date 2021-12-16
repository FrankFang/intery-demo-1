package deploy

import (
	"fmt"
	"intery/server/database"
	"intery/server/engine/base"
	"intery/server/model"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/go-github/v41/github"
)

type Controller struct {
	base.BaseController
}

func (ctrl *Controller) Create(c *gin.Context) {
	var params struct {
		ProjectId string `json:"project_id" binding:"required"`
	}
	if err := c.BindJSON(&params); err != nil {
		log.Fatal(err)
	}

	project := &model.Project{}
	database.GetDB().First(project, "id = ?", params.ProjectId)
	user, auth, err := ctrl.GetUserAndAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": err.Error()})
		return
	}
	if project.UserId != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "该项目不属于你"})
		return
	}
	// download code from github
	client := ctrl.GetGithubClient(c, auth)
	_, dir, response, err := client.Repositories.GetContents(c, auth.Login, project.RepoName, "/", &github.RepositoryContentGetOptions{
		Ref: "main",
	})
	fmt.Println(dir, response, err)

	// save code to file system
	// run docker

}
