package deploy

import (
	"fmt"
	"intery/cmd/docker"
	"intery/server/config/dir"
	"intery/server/database"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ahmetb/dlog"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

type Options struct {
	AppKind        string
	ContainerName  string
	ProjectDir     string
	SocketDir      string
	SocketFileName string
	Path           string
}

func RemoveCurrentContainer(c *gin.Context, deploymentId uint) error {
	d := database.GetQuery().Deployment
	deployment, err := d.WithContext(c).Where(d.ID.Eq(deploymentId)).First()
	if err != nil {
		return err
	}
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
	if deployment.Status == "running" {
		logReader, err := cli.ContainerLogs(c, deployment.ContainerId, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
			Tail:       "10000",
		})
		if err != nil {
			log.Println("Get container logs failed. ", err)
			return err
		}
		userDir := dir.EnsureUserDir(deployment.UserId)
		projectDir := dir.EnsureProjectDir(userDir, deployment.ProjectId)

		r := dlog.NewReader(logReader)
		file, err := os.Create(filepath.Join(projectDir, deployment.ContainerId+"_log"))

		if err != nil {
			log.Println("Create file failed. ", err)
			return err
		}
		defer file.Close()
		go func() {
			time.Sleep(100 * time.Millisecond)
			logReader.Close()
		}()
		io.Copy(file, r)
	}
	cli.ContainerRemove(c, deployment.ContainerId, types.ContainerRemoveOptions{
		Force: true,
	})
	return nil
}

func CreateAndStartContainer(c *gin.Context, opt Options) (string, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}
	var imageName string
	var command string
	var memory int64
	if opt.AppKind == "nodejs" {
		imageName = "node:16.13"
		command = "npm i --registry=https://registry.npmmirror.com && node server.js"
		memory = 128 * 1024 * 1024
	} else if opt.AppKind == "golang" {
		imageName = "golang:1.17.6"
		command = "go env -w GOPROXY=https://goproxy.cn,direct && go run main.go"
		memory = 512 * 1024 * 1024
	}

	config := container.Config{
		Image:      imageName,
		WorkingDir: "/app",
		Cmd:        []string{"/bin/sh", "-c", command},
		// Cmd: []string{"/usr/local/bin/node", "server.js"},
		Env: []string{
			fmt.Sprintf("SOCKET=/tmp/socket/%s", opt.SocketFileName),
		},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	}
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Target: "/app/",
				Source: opt.Path,
			},
			{
				Type:     "bind",
				Target:   "/tmp/socket/",
				Source:   opt.SocketDir,
				ReadOnly: false,
			},
		},
		Resources: container.Resources{
			Memory: memory,
		},
	}
	container, err := cli.ContainerCreate(c, &config, &hostConfig, nil, nil, opt.ContainerName)
	if err != nil {
		return container.ID, err
	}
	if err = cli.ContainerStart(c, container.ID, types.ContainerStartOptions{}); err != nil {
		return container.ID, err
	}
	err = docker.ReloadNginx(c)
	return container.ID, err
}
