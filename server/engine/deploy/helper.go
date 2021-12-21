package deploy

import (
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/gin-gonic/gin"
)

type Options struct {
	ImageName     string
	ContainerName string
	Port          string
	Path          string
}

func CreateDockerContainer(ctx *gin.Context, opt Options) error {

	// 创建客户端
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal(err)
	}

	config := container.Config{
		Image:      opt.ImageName,
		WorkingDir: "/app",
		Cmd:        []string{"/usr/local/bin/node", "server.js"},
		Env: []string{
			"NODE_ENV=production",
		},
		ExposedPorts: nat.PortSet{
			"8080/tcp": struct{}{},
		},
	}
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Target: "/app/",
				Source: opt.Path,
			},
		},
		PortBindings: nat.PortMap{
			"8080/tcp": []nat.PortBinding{
				{
					HostIP:   "",
					HostPort: opt.Port,
				},
			},
		},
	}
	// 用客户端创建容器
	body, err := cli.ContainerCreate(ctx, &config, &hostConfig, nil, nil, opt.ContainerName)
	if err != nil {
		log.Fatal(err)
	}
	// 以 -d 选项启动容器
	if err := cli.ContainerStart(ctx, body.ID, types.ContainerStartOptions{}); err != nil {
		log.Fatal(err)
	}
	return nil
}
