package deploy

import (
	"fmt"
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
	ImageName      string
	ContainerName  string
	ProjectDir     string
	SocketDir      string
	SocketFileName string
	Path           string
}

func RemoveCurrentContainer(c *gin.Context, deploymentId uint) error {
	d := database.GetQuery().Deployment
	deployment, err := d.WithContext(c).Where(d.ID.Eq(deploymentId)).Where(d.Status.Eq("running")).First()
	if err != nil {
		return err
	}

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}
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
	err = cli.ContainerRemove(c, deployment.ContainerId, types.ContainerRemoveOptions{
		Force: true,
	})
	return err
}

func CreateAndStartNodejsContainer(c *gin.Context, opt Options) (string, error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return "", err
	}

	config := container.Config{
		Image:      opt.ImageName,
		WorkingDir: "/app",
		// Cmd:        []string{"/bin/sh", "-c", "echo fuck > /tmp/log; /usr/local/bin/node server.js 2>&1 >> /tmp/log"},
		Cmd: []string{"/usr/local/bin/node", "server.js"},
		Env: []string{
			fmt.Sprintf("PORT=/tmp/socket/%s", opt.SocketFileName),
			"NODE_ENV=production",
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
	}
	body, err := cli.ContainerCreate(c, &config, &hostConfig, nil, nil, opt.ContainerName)
	if err != nil {
		return body.ID, err
	}
	if err = cli.ContainerStart(c, body.ID, types.ContainerStartOptions{}); err != nil {
		return body.ID, err
	}
	return body.ID, err
}
