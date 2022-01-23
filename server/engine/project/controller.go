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
	if len(projects) > 3 {
		c.JSON(http.StatusTooManyRequests, gin.H{"reason": "目前每个账户只能创建 3 个项目"})
		return
	}
	oauth2Token := oauth2.Token{AccessToken: auth.AccessToken, RefreshToken: "hi"}
	client := sdk.NewClient(github.Conf.Client(c, &oauth2Token))
	repo, _, err := client.Repositories.Create(c, "", &sdk.Repository{
		Name:    sdk.String(params.RepoName),
		Private: sdk.Bool(true),
	})
	if err != nil {
		ctrl.HandleGitHubError(c, err)
		log.Println("Create repo failed.", err)
		return
	}
	repoContent, _, err := client.Repositories.CreateFile(c, auth.Login, params.RepoName, "README.md", &sdk.RepositoryContentFileOptions{
		Content: []byte("# " + params.RepoName),
		Message: sdk.String("Initial commit"),
	})
	if err != nil {
		ctrl.HandleGitHubError(c, err)
		log.Println("Create file failed.", err)
		return
	}
	tree, _, err := client.Git.GetTree(c, auth.Login, params.RepoName, *repoContent.SHA, true)
	if err != nil {
		ctrl.HandleGitHubError(c, err)
		log.Println("Get tree failed.", err)
		return
	}
	files := getAppFiles(params.AppKind, struct{ Name string }{Name: params.RepoName})
	fileTree := make([]*sdk.TreeEntry, 0, 128)
	for _, file := range files {
		fileTree = append(fileTree, &sdk.TreeEntry{
			Path:    sdk.String(file.Path),
			Mode:    sdk.String("100644"),
			Type:    sdk.String("blob"),
			Content: sdk.String(file.Content),
		})
	}
	newTree, _, err := client.Git.CreateTree(c, auth.Login, params.RepoName, *tree.SHA, fileTree)
	if err != nil {
		ctrl.HandleGitHubError(c, err)
		log.Println("Create tree failed.", err)
		return
	}
	newCommit, _, err := client.Git.CreateCommit(c, auth.Login, params.RepoName, &sdk.Commit{
		Message: sdk.String("Second commit"),
		Tree:    newTree,
		Parents: []*sdk.Commit{
			{
				SHA: tree.SHA,
			},
		},
	})
	if err != nil {
		ctrl.HandleGitHubError(c, err)
		log.Println("Create commit failed.", err)
		return
	}
	_, _, err = client.Git.UpdateRef(c, auth.Login, params.RepoName, &sdk.Reference{
		Ref: sdk.String("refs/heads/main"),
		Object: &sdk.GitObject{
			SHA: newCommit.SHA,
		},
	}, false)
	if err != nil {
		ctrl.HandleGitHubError(c, err)
		log.Println("Update ref failed.", err)
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
