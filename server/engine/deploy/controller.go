package deploy

import (
	"fmt"
	"intery/cmd/docker"
	"intery/server/database"
	"intery/server/engine/base"
	"intery/server/model"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	p := database.GetQuery().Project
	project, err := p.WithContext(c).Where(p.ID.Eq(params.ProjectId)).First()
	if err != nil {
		log.Println(err)
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
		ctrl.HandleGitHubError(c, err)
		log.Println("Get archive link failed. ", err)
		return
	}
	cwd, _ := os.Getwd()
	userDir := filepath.Join(cwd, "/userspace/", strconv.Itoa(int(user.ID)))
	if err := os.RemoveAll(userDir); err != nil {
		log.Println("Remove userDir failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if err := os.MkdirAll(userDir, os.ModePerm); err != nil {
		log.Println("Make userDir failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	projectDir := filepath.Join(userDir, strconv.Itoa(int(project.ID)))
	if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
		log.Println("Make projectDir failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	archivePath := filepath.Join(projectDir, "src.zip")
	srcDir := filepath.Join(projectDir, "src")
	socketDir := filepath.Join(cwd, "userspace", "socket")
	confPath := filepath.Join(cwd, "config", "nginx_default.conf")
	if err := os.MkdirAll(socketDir, os.ModePerm); err != nil {
		log.Println("Make socketDir failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if err := os.Chmod(socketDir, 0777); err != nil {
		log.Println("Chmod socketDir failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if err := downloadFile(url.String(), archivePath); err != nil {
		log.Println("Download url failed. ", url.String(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	uz := unzip.New(archivePath, srcDir)
	err = uz.Extract()
	if err != nil {
		log.Println("Extract archive failed. ", archivePath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	var dirName string
	if files, err := ioutil.ReadDir(srcDir); err != nil {
		log.Println("Read srcDir failed. ", srcDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	} else {
		for _, node := range files {
			if node.IsDir() {
				dirName = node.Name()
				break
			}
		}
	}
	socketFileName := fmt.Sprintf("%s.sock", strconv.Itoa(int(project.ID)))
	if err := os.RemoveAll(filepath.Join(socketDir, socketFileName)); err != nil {
		log.Println("Remove socketDir failed. ", socketDir, err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	containerId, err := CreateAndStartNodejsContainer(c, Options{
		ImageName:      "node:latest",
		ContainerName:  fmt.Sprintf("app_%d_%d", user.ID, project.ID),
		SocketDir:      socketDir,
		SocketFileName: socketFileName,
		Path:           filepath.Join(srcDir, dirName),
	})
	if err != nil {
		log.Println("Create container failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "创建容器失败"})
		return
	}
	comments := "# Append content below"
	content, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Println("Read confPath failed. ", confPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	s := string(content)
	if !strings.Contains(s, fmt.Sprintf("location /%d/", project.ID)) {
		s = strings.Replace(s, comments, comments+"\n"+fmt.Sprintf(`
	location /%d/ {
		proxy_pass http://upstream_%d;
		proxy_set_header            Host $host;
		proxy_set_header            X-Real-IP $remote_addr;
		proxy_http_version          1.1;
		proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
		proxy_set_header            X-Forwarded-Proto http;
		proxy_redirect              http:// $scheme://;
	}
`, project.ID, project.ID), -1)
		s = fmt.Sprintf(`upstream upstream_%d {
	server unix:/tmp/socket/%d.sock;
}
`, project.ID, project.ID) + s
	}
	err = ioutil.WriteFile(confPath, []byte(s), 0777)
	if err != nil {
		log.Println(err)
	}
	err = docker.ReloadNginx(c)
	if err != nil {
		log.Println("Reload nginx failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	d := database.GetQuery().Deployment
	deployment := model.Deployment{
		ProjectId:   project.ID,
		ContainerId: containerId,
	}
	err = d.WithContext(c).Create(&deployment)
	if err != nil {
		log.Println("Save deployment failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	_, err = p.WithContext(c).Where(p.ID.Eq(project.ID)).
		Update(p.LatestDeploymentId, deployment.ID)
	if err != nil {
		log.Println("Update project failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"resource": deployment})
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
