package deploy

import (
	"fmt"
	"intery/cmd/docker"
	"intery/lib/unzip"
	"intery/server/config/dir"
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
	"time"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	base.BaseController
}

func (ctrl *Controller) Index(c *gin.Context) {
	page, perPage, offset, err := ctrl.MustHasPage(c)
	if err != nil {
		return
	}
	auth, err := ctrl.MustAuth(c)
	if err != nil {
		return
	}
	projectIdString := c.Query("project_id")
	projectId, err := strconv.Atoi(projectIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "project_id 必须是一个数字"})
		return
	}
	db := database.GetDB()
	q := db.Model(model.Deployment{}).
		Where("user_id = ?", auth.UserId).
		Where("project_id = ?", projectId).
		Order("created_at desc")
	d := database.GetQuery().Deployment
	query := d.WithContext(c).Where(d.UserId.Eq(auth.UserId)).Where(d.ProjectId.Eq(uint(projectId))).Order(d.CreatedAt.Desc())
	deployments, err := query.Offset(offset + perPage*(page-1)).Limit(perPage).Debug().Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	var count int64
	q.Count(&count)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"resource": deployments,
		"pager": gin.H{
			"count":    count,
			"per_page": perPage,
			"page":     page,
		},
	})
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
	userDir := dir.EnsureUserDir(user.ID)
	if err != nil {
		log.Println("Make userDir failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	projectDir := dir.EnsureProjectDir(userDir, project.ID)
	if err := os.MkdirAll(projectDir, os.ModePerm); err != nil {
		log.Println("Make projectDir failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	archivePath := filepath.Join(projectDir, fmt.Sprintf("src-%d.zip", time.Now().Unix()))
	srcDir := filepath.Join(projectDir, "src")
	socketDir := dir.GetSocketDir()
	confPath := filepath.Join(dir.GetNginxConfigDir(), "default.conf")
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

	uz := unzip.New()
	_, err = uz.Extract(archivePath, srcDir)
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
	d := database.GetQuery().Deployment

	if project.LatestDeploymentId != 0 {
		err = RemoveCurrentContainer(c, project.LatestDeploymentId)
		if err != nil {
			log.Println("Remove current container failed. ", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"reason":        err.Error(),
				"deployment_id": project.LatestDeploymentId,
			})
			return
		}
	}
	_, err = d.WithContext(c).Where(d.ID.Eq(project.LatestDeploymentId)).UpdateColumn(d.Status, "removed")
	if err != nil {
		log.Println("Update deployment's status failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}

	containerId, err := CreateAndStartContainer(c, Options{
		AppKind:        project.AppKind,
		ContainerName:  fmt.Sprintf("app_%d_%d", user.ID, project.ID),
		ProjectDir:     projectDir,
		SocketDir:      socketDir,
		SocketFileName: socketFileName,
		Path:           filepath.Join(srcDir, dirName),
	})
	if err != nil {
		log.Println("Create container failed. ", err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": "创建容器失败"})
		return
	}
	comments := "# Placeholder"
	content, err := ioutil.ReadFile(confPath)
	if err != nil {
		log.Println("Read confPath failed. ", confPath, err)
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	s := string(content)
	if !strings.Contains(s, fmt.Sprintf("location /preview/%d/", project.ID)) {
		s = strings.Replace(s, comments, comments+"\n"+fmt.Sprintf(`
  location /preview/%d/ {
    proxy_pass http://unix:/tmp/socket/%d.sock:/;
    proxy_set_header            Host $host;
    proxy_set_header            X-Real-IP $remote_addr;
    proxy_http_version          1.1;
    proxy_set_header            X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header            X-Forwarded-Proto http;
    proxy_redirect              http:// $scheme://;
	}
`, project.ID, project.ID), -1)
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
	deployment := model.Deployment{
		ProjectId:   project.ID,
		ContainerId: containerId,
		UserId:      auth.UserId,
		Status:      "running",
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

func (ctrl *Controller) Show(c *gin.Context) {
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": "id is not a number"})
		return
	}
	d := database.GetQuery().Deployment
	deployment, err := d.WithContext(c).Where(d.ID.Eq(uint(id))).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": err.Error()})
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
