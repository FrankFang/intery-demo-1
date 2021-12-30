package deploy

import (
	"fmt"
	"os/exec"

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
		Cmd:        []string{"/bin/sh", "-c", "echo fuck > /tmp/log; /usr/local/bin/node server.js 2>&1 >> /tmp/log"},
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
	body, err := cli.ContainerCreate(ctx, &config, &hostConfig, nil, nil, opt.ContainerName)
	if err != nil {
		return body.ID, err
	}
	if err = cli.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		return body.ID, err
	}
	// execute "docker logs"
	cmd := exec.Command("docker", "logs", body.ID)
	stdout, cmderr := cmd.Output()
	if cmderr != nil {
		fmt.Println(cmderr.Error())
	}
	fmt.Print(stdout)
	return body.ID, err
}
