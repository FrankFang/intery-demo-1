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
	cwd, _ := os.Getwd()
	userDir := filepath.Join(cwd, "/userspace/", strconv.Itoa(int(user.ID)))
	if err := os.RemoveAll(userDir); err != nil {
		log.Fatal(err)
	}
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	projectDir := filepath.Join(userDir, strconv.Itoa(int(project.ID)))
	if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	archivePath := filepath.Join(projectDir, "src.zip")
	srcDir := filepath.Join(projectDir, "src")
	socketDir := filepath.Join(projectDir, "socket")
	if err := os.MkdirAll(socketDir, os.ModePerm); err != nil {
		log.Fatal(err)
	}
	if err := os.Chmod(socketDir, 0777); err != nil {
		log.Fatal(err)
	}
	if err := downloadFile(url.String(), archivePath); err != nil {
		log.Fatal(err)
	}

	uz := unzip.New(archivePath, srcDir)
	err = uz.Extract()
	if err != nil {
		fmt.Println(err)
	}
	var dirName string
	if files, err := ioutil.ReadDir(srcDir); err != nil {
		log.Fatal(err)
	} else {
		for _, node := range files {
			if node.IsDir() {
				dirName = node.Name()
				break
			}
		}
	}
	containerId, err := CreateDockerContainer(c, Options{
		ImageName: "node:latest",
		SocketDir: socketDir,
		Path:      filepath.Join(srcDir, dirName),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	// reader, _ := GetContainerLogs(c, containerId)
	// defer reader.Close()
	// logs, _ := ioutil.ReadAll(reader)
	// c.JSON(http.StatusOK, gin.H{"resource": gin.H{"containerId": containerId}, "logs": logs})
	c.JSON(http.StatusOK, gin.H{"resource": gin.H{"containerId": containerId}})
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
