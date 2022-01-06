package project

import (
	"bytes"
	"encoding/json"
	"html/template"
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
		log.Fatal(err)
	}
	user, auth, err := ctrl.GetUserAndAuth(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"reason": err.Error()})
		return
	}
	oauth2Token := oauth2.Token{AccessToken: auth.AccessToken, RefreshToken: "hi"}
	client := sdk.NewClient(github.Conf.Client(c, &oauth2Token))
	repo, _, err := client.Repositories.Create(c, "", &sdk.Repository{
		Name:    sdk.String(params.RepoName),
		Private: sdk.Bool(true),
	})
	if err != nil {
		if resErr, ok := err.(*sdk.ErrorResponse); ok {
			defer resErr.Response.Body.Close()
			transformResponse(c, resErr.Response)
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"reason": err.Error()})
			log.Fatal("Create repo failed.", err)
		}
		return
	}
	repoContent, _, err := client.Repositories.CreateFile(c, auth.Login, params.RepoName, "README.md", &sdk.RepositoryContentFileOptions{
		Content: []byte("# " + params.RepoName),
		Message: sdk.String("Initial commit"),
	})
	if err != nil {
		if err, ok := err.(*sdk.ErrorResponse); ok {
			defer err.Response.Body.Close()
			transformResponse(c, err.Response)
			return
		} else {
			log.Fatal("Create file failed.", err)
		}
	}
	tree, _, err := client.Git.GetTree(c, auth.Login, params.RepoName, *repoContent.SHA, true)
	if err != nil {
		if err, ok := err.(*sdk.ErrorResponse); ok {
			defer err.Response.Body.Close()
			transformResponse(c, err.Response)
			return
		} else {
			log.Fatal("Get tree failed.", err)
		}
	}
	files := getNodejsAppFiles(struct{ Name string }{Name: params.RepoName})
	fileTree := make([]*sdk.TreeEntry, 0, 128)
	for _, file := range files {
		fileTree = append(fileTree, &sdk.TreeEntry{
			Path:    sdk.String(file.Path),
			Mode:    sdk.String("100644"),
			Type:    sdk.String("blob"),
			Content: sdk.String(file.Content),
		})
	}
	newTree, _, _ := client.Git.CreateTree(c, auth.Login, params.RepoName, *tree.SHA, fileTree)
	newCommit, _, _ := client.Git.CreateCommit(c, auth.Login, params.RepoName, &sdk.Commit{
		Message: sdk.String("Second commit"),
		Tree:    newTree,
		Parents: []*sdk.Commit{
			{
				SHA: tree.SHA,
			},
		},
	})
	_, _, _ = client.Git.UpdateRef(c, auth.Login, params.RepoName, &sdk.Reference{
		Ref: sdk.String("refs/heads/main"),
		Object: &sdk.GitObject{
			SHA: newCommit.SHA,
		},
	}, false)
	// create project and save to database
	project := model.Project{
		AppKind:  params.AppKind,
		RepoName: params.RepoName,
		UserId:   user.ID,
		RepoHome: repo.GetHTMLURL(),
	}
	err = database.GetQuery().Project.WithContext(c).Create(&project)
	if err != nil {
		log.Fatal(err)
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
	query := p.WithContext(c).Where(p.UserId.Eq(auth.UserId))
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

// helper function
type Node struct {
	Path    string
	Content string
}

func getNodejsAppFiles(data interface{}) (nodes []Node) {
	currentDir, _ := os.Getwd()
	// FIXME: hard code
	if gin.Mode() == gin.TestMode {
		for !strings.HasSuffix(currentDir, "intery-demo-1" /*project dir name*/) {
			currentDir = filepath.Dir(currentDir)
		}
	}
	dir := filepath.Join(currentDir, "server/app_templates/nodejs")
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		content, _ := ioutil.ReadFile(path)
		t, _ := template.New("text").Parse(string(content))
		var b bytes.Buffer
		t.Execute(&b, data)
		relativePath, _ := filepath.Rel(dir, path)
		nodes = append(nodes, Node{
			Path:    relativePath,
			Content: b.String(),
		})
		return nil
	})
	return
}

func transformResponse(c *gin.Context, response *http.Response) {
	c.Status(response.StatusCode)
	if response.ContentLength == 0 {
		return
	}
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal("Read response body failed.", err)
	}
	var data struct {
		DocumentationUrl string `json:"documentation_url"`
		Errors           []struct {
			Code     string
			Field    string
			Message  string
			Resource string
		}
		Message string `json:"message"`
	}
	json.Unmarshal(content, &data)

	if data.Message == "Repository creation failed." {
		c.JSON(http.StatusUnprocessableEntity, gin.H{"errors": gin.H{
			"repo_name": []string{"GitHub上已经存在该仓库，请使用其他仓库名。"},
		}})
		return
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"reason": data.Message})
	}
}
