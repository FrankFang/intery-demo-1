package deploy

import (
	"fmt"
	"io"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/gin-gonic/gin"
)

type Options struct {
	ImageName      string
	ContainerName  string
	SocketDir      string
	SocketFileName string
	Path           string
}

func CreateDockerContainer(ctx *gin.Context, opt Options) (containerId string, err error) {

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}
	_, err = cli.ImagePull(ctx, opt.ImageName, types.ImagePullOptions{})
	if err != nil {
		return
	}

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: opt.ContainerName}),
		All:     true,
	})
	if err != nil {
		return
	}
	if len(containers) > 0 {
		containerId = containers[0].ID
		err = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
			Force: true,
		})
		if err != nil {
			return
		}
	}

	config := container.Config{
		Image:      opt.ImageName,
		WorkingDir: "/app",
		Cmd:        []string{"/usr/local/bin/node", "server.js"},
		Env: []string{
			fmt.Sprintf("PORT=/tmp/socket/%s", opt.SocketFileName),
			"NODE_ENV=production",
		},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
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
	body, err := cli.ContainerCreate(ctx, &config, &hostConfig, nil, nil, opt.ContainerName)
	if err != nil {
		return body.ID, err
	}
	if err = cli.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		return body.ID, err
	}
	return body.ID, err
}

func GetContainerLogs(ctx *gin.Context, containerId string) (reader io.ReadCloser, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}
	reader, err = cli.ContainerLogs(ctx, containerId, types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
	})
	return
}
