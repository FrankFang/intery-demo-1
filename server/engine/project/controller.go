package project

import (
	"bytes"
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
	sdk "github.com/google/go-github/v41/github"
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
	client := sdk.NewClient(github.Conf.Client(c, &oauth2Token))

	repo, _, err := client.Repositories.CreateFromTemplate(c, "jirengu-inc", "intery-template-"+params.AppKind, &sdk.TemplateRepoRequest{
		Name: sdk.String(params.RepoName),
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

	project := model.Project{
		AppKind:  params.AppKind,
		RepoName: params.RepoName,
		UserId:   user.ID,
		RepoHome: repo.GetHTMLURL(),
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
