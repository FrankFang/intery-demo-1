package project

import (
	"bytes"
	"encoding/json"
	"fmt"
	"html/template"
	"intery/server/config/dir"
	"intery/server/database"
	"intery/server/engine/auth/github"
	"intery/server/model"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"intery/server/engine/base"

	"github.com/gin-gonic/gin"
	githubSdk "github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type Controller struct {
	base.BaseController
}

func (ctrl *Controller) Create(c *gin.Context) {
	var params struct {
		AppKind  string `json:"app_kind" binding:"required"`
		RepoName string `json:"repo_name" binding:"required"`
	}
	if err := c.BindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}
	params.RepoName = params.AppKind + "-" + params.RepoName
	user, auth, err := ctrl.GetUserAndAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": err.Error()})
		return
	}
	p := database.GetQuery().Project
	projects, err := p.WithContext(c).Where(p.UserId.Eq(user.ID)).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if len(projects) >= 3 {
		c.JSON(http.StatusTooManyRequests, gin.H{"reason": "目前每个账户只能创建 3 个项目"})
		return
	}
	oauth2Token := oauth2.Token{AccessToken: auth.AccessToken, RefreshToken: "none"}
	var repoHome string
	if auth.Provider == "github" {
		githubClient := githubSdk.NewClient(github.Conf.Client(c, &oauth2Token))

		repo, _, err := githubClient.Repositories.CreateFromTemplate(c, "jirengu-inc", "intery-template-"+params.AppKind, &githubSdk.TemplateRepoRequest{
			Name: githubSdk.String(params.RepoName),
		})

		if err != nil {
			if i := strings.Index(err.Error(), "Name already exists"); i >= 0 {
				c.JSON(http.StatusUnprocessableEntity, gin.H{
					"errors": gin.H{
						"repo_name": []string{"GitHub上已经存在该仓库，请使用其他仓库名。"},
					},
				})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
			}
			return
		}
		repoHome = repo.GetHTMLURL()
	} else {
		h := gin.H{"access_token": auth.AccessToken}
		jsonData, _ := json.Marshal(h)
		_, err := http.Post(
			fmt.Sprintf("https://gitee.com/api/v5/repos/jirengu-inc/intery-template-%s/forks", params.AppKind),
			"application/json;charset=UTF-8",
			bytes.NewBuffer(jsonData),
		)
		if err != nil {
			log.Println("Create gitee repo failed. ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
			return
		}
		h = gin.H{
			"name": params.RepoName, "owner": auth.Login,
			"repo": params.RepoName, "access_token": auth.AccessToken,
			"path": params.RepoName,
		}
		jsonData, _ = json.Marshal(h)
		req, _ := http.NewRequest("PATCH",
			fmt.Sprintf("https://gitee.com/api/v5/repos/%s/%s", auth.Login, "intery-template-"+params.AppKind),
			bytes.NewBuffer(jsonData),
		)
		req.Header.Set("Content-Type", "application/json;charset=UTF-8")
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println("Do request failed. ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println("Read response body failed. ", err)
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err})
			return
		}
		var tmp struct {
			HtmlUrl string `json:"html_url"`
		}
		json.Unmarshal(body, &tmp)
		repoHome = tmp.HtmlUrl
	}

	project := model.Project{
		AppKind:  params.AppKind,
		RepoName: params.RepoName,
		UserId:   user.ID,
		RepoHome: repoHome,
	}
	err = database.GetQuery().Project.WithContext(c).Create(&project)
	if err != nil {
		log.Println(err)
	}
	c.JSON(http.StatusCreated, gin.H{"resource": project})
}

func (ctrl *Controller) Show(c *gin.Context) {
	_, err := ctrl.MustSignIn(c)
	if err != nil {
		return
	}
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}
	p := database.GetQuery().Project
	project, err := p.WithContext(c).Where(p.ID.Eq(uint(id))).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"resource": project})
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
	p := database.GetQuery().Project
	query := p.WithContext(c).Where(p.UserId.Eq(auth.UserId)).Order(p.CreatedAt.Desc())
	projects, err := query.Offset(offset + perPage*(page-1)).Limit(perPage + 1).Find()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	has_next_page := len(projects) > perPage
	if has_next_page {
		projects = projects[:len(projects)-1]
	}
	count, err := query.Count()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"resource": projects,
		"pager": gin.H{
			"count":         count,
			"per_page":      perPage,
			"page":          page,
			"has_next_page": has_next_page,
		},
	})
}
func (ctrl *Controller) Delete(c *gin.Context) {
	_, err := ctrl.MustSignIn(c)
	if err != nil {
		return
	}
	idString := c.Param("id")
	id, err := strconv.Atoi(idString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"reason": err.Error()})
		return
	}
	p := database.GetQuery().Project
	project, err := p.WithContext(c).Where(p.ID.Eq(uint(id))).First()
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"reason": err.Error()})
		return
	}
	_, err = p.WithContext(c).Where(p.ID.Eq(uint(id))).Update(p.DeletedAt, time.Now())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
		return
	}
	project.DeletedAt.Time = time.Now()
	c.JSON(http.StatusOK, gin.H{"resource": project})
}

// helper function
type Node struct {
	Path    string
	Content string
}

func getAppFiles(appKind string, data interface{}) (nodes []Node) {
	cwd, _ := os.Getwd()
	// FIXME: hard code
	if gin.Mode() == gin.TestMode {
		for !strings.HasSuffix(cwd, "intery-demo-1" /*project dir name*/) {
			cwd = filepath.Dir(cwd)
		}
	}
	d := dir.GetAppTemplatesDir(appKind)
	err := filepath.Walk(d, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			log.Println(err)
			return err
		}
		if f.IsDir() {
			return nil
		}
		content, err := ioutil.ReadFile(path)
		if err != nil {
			log.Println(err)
			return err
		}
		t, err := template.New("text").Parse(string(content))
		if err != nil {
			log.Println(err)
			return err
		}
		var b bytes.Buffer
		t.Execute(&b, data)
		relativePath, err := filepath.Rel(d, path)
		if err != nil {
			log.Println(err)
			return err
		}
		nodes = append(nodes, Node{
			Path:    relativePath,
			Content: b.String(),
		})
		return nil
	})
	if err != nil {
		log.Println(err)
	}
	return
}
