package deploy

import (
	"fmt"
	"intery/server/database"
	"intery/server/engine/base"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/artdarek/go-unzip"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	base.BaseController
}

func (ctrl *Controller) Create(c *gin.Context) {
	var params struct {
		ProjectId uint `json:"project_id" binding:"required"`
	}
	if err := c.BindJSON(&params); err != nil {
		log.Fatal(err)
	}
	p := database.GetQuery().Project
	project, err := p.WithContext(c).Where(p.ID.Eq(params.ProjectId)).First()
	if err != nil {
		log.Fatal(err)
	}
	user, auth, err := ctrl.GetUserAndAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": err.Error()})
		return
	}
	if project.UserId != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"reason": "该项目不属于该用户"})
		return
	}
	client := ctrl.GetGithubClient(c, auth)

	// download github repo as a archive file
	url, _, err := client.Repositories.GetArchiveLink(c, auth.Login, project.RepoName, "zipball", nil, false)
	if err != nil {
		log.Fatal(err)
	}
	dir := filepath.Join("/root/intery-userspace/", strconv.Itoa(int(user.ID)), project.RepoName)
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	archivePath := filepath.Join(dir, "archive.zip")
	targetPath := filepath.Join(dir, "src")
	err = downloadFile(url.String(), archivePath)
	if err != nil {
		log.Fatal(err)
	}

	uz := unzip.New(archivePath, targetPath)
	err = uz.Extract()
	if err != nil {
		fmt.Println(err)
	}
	var dirName string
	files, err := ioutil.ReadDir(targetPath)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		if file.IsDir() {
			dirName = file.Name()
			break
		}
	}
	fmt.Println(dirName)
	err = CreateDockerContainer(c, Options{
		ImageName: "node:latest",
		Port:      "7777",
		Path:      filepath.Join(targetPath, dirName),
	})
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(http.StatusOK, gin.H{"resource": gin.H{"url": "http://localhost:7777"}})
}

func downloadFile(url string, filepath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, resp.Body)
	return err
}
