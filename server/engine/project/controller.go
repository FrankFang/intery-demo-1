package project

import (
	"bytes"
	"crypto/x509"
	"encoding/pem"
	"html/template"
	"intery/server/database"
	"intery/server/engine/auth/github"
	"intery/server/models"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	sdk "github.com/google/go-github/v41/github"
	"golang.org/x/oauth2"
)

type Controller struct {
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
	c.Writer.Write(content)
}
func getUserIdFromHeader(c *gin.Context) (userId uint, err error) {
	bearer := c.Request.Header.Get("Authorization")
	parts := strings.Split(bearer, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		err = errors.New("invalid authorization")
		return
	}
	jwtString := parts[1]
	token, err := jwt.Parse(jwtString, keyFunc)
	if err != nil {
		err = errors.New("invalid jwt")
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		err = errors.New("invalid claims")
		return
	}
	userId = uint(claims["user_id"].(float64))
	return
}

func getAuthFromUserId(userId uint) (auth *models.Authorization) {
	database.GetDB().Find(&auth, "user_id = ?", userId)
	return auth
}

func (ctrl *Controller) Create(c *gin.Context) {
	var params struct {
		AppKind  string `json:"app_kind" binding:"required"`
		RepoName string `json:"repo_name" binding:"required"`
	}
	if err := c.BindJSON(&params); err != nil {
		log.Fatal(err)
	}
	userId, err := getUserIdFromHeader(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
	}
	auth := getAuthFromUserId(userId)
	oauth2Token := oauth2.Token{AccessToken: auth.AccessToken, RefreshToken: "hi"}
	client := sdk.NewClient(github.Conf.Client(c, &oauth2Token))
	repo, _, err := client.Repositories.Create(c, "", &sdk.Repository{
		Name:    sdk.String(params.RepoName),
		Private: sdk.Bool(false),
	})
	if err != nil {
		if err, ok := err.(*sdk.ErrorResponse); ok {
			defer err.Response.Body.Close()
			transformResponse(c, err.Response)
			return
		} else {
			log.Fatal("Create repo failed.", err)
		}
	}
	repoContent, _, _ := client.Repositories.CreateFile(c, auth.Login, repo.GetName(), "README.md", &sdk.RepositoryContentFileOptions{
		Content: []byte("# " + repo.GetName()),
		Message: sdk.String("Initial commit"),
	})
	tree, _, _ := client.Git.GetTree(c, auth.Login, repo.GetName(), *repoContent.SHA, true)
	files := getNodejsAppFiles(struct{ Name string }{Name: repo.GetName()})
	fileTree := make([]*sdk.TreeEntry, 0, 128)
	for _, file := range files {
		fileTree = append(fileTree, &sdk.TreeEntry{
			Path:    sdk.String(file.Path),
			Mode:    sdk.String("100644"),
			Type:    sdk.String("blob"),
			Content: sdk.String(file.Content),
		})
	}
	newTree, _, _ := client.Git.CreateTree(c, auth.Login, repo.GetName(), *tree.SHA, fileTree)
	newCommit, _, _ := client.Git.CreateCommit(c, auth.Login, repo.GetName(), &sdk.Commit{
		Message: sdk.String("Second commit"),
		Tree:    newTree,
		Parents: []*sdk.Commit{
			{
				SHA: tree.SHA,
			},
		},
	})
	_, _, _ = client.Git.UpdateRef(c, auth.Login, repo.GetName(), &sdk.Reference{
		Ref: sdk.String("refs/heads/main"),
		Object: &sdk.GitObject{
			SHA: newCommit.SHA,
		},
	}, false)
	// create project and save to database
	project := models.Project{
		AppKind:  params.AppKind,
		RepoName: repo.GetName(),
		UserId:   userId,
		RepoHome: repo.GetHTMLURL(),
	}
	err = project.Create()
	if err != nil {
		log.Fatal(err)
	}
	c.JSON(http.StatusCreated, project)
}

// helper function
func keyFunc(t *jwt.Token) (interface{}, error) {
	if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
		log.Fatal("method not supported")
	}
	key, err := ioutil.ReadFile(os.Getenv("PUBLIC_KEY"))
	if err != nil {
		log.Fatal(err)
	}
	block, _ := pem.Decode(key)
	if block == nil || block.Type != "PUBLIC KEY" {
		log.Fatal("failed to decode PEM block containing public key")
	}

	pub, err := x509.ParsePKIXPublicKey(block.Bytes)
	if err != nil {
		log.Fatal(err)
	}
	return pub, nil
}

type Node struct {
	Path    string
	Content string
}

func getNodejsAppFiles(data interface{}) (nodes []Node) {
	currentDir, _ := os.Getwd()
	dir := filepath.Join(currentDir, "server/app_templates/nodejs")
	// walk through dir and filter all files's relative path to dir
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f.IsDir() {
			return nil
		}
		// get content from path
		content, _ := ioutil.ReadFile(path)
		// use template to parse the content
		t, _ := template.New("text").Parse(string(content))
		// t execute and get the result to x
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