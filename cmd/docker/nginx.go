package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

const NginxContainerName = "nginx1"

func StartNginx() (containerId string, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}

	ctx := context.Background()
	containerId, err = getNginxContainerId(ctx, cli)
	if err != nil {
		return
	}
	if containerId == "" {
		fmt.Println("containerId is empty")
		containerId, err = createNginxContainer(ctx, cli)
		if err != nil {
			return
		}
	}

	inspect, err := cli.ContainerInspect(ctx, containerId)
	if !inspect.State.Running {
		err = cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
		if err != nil {
			return
		}
	}
	return
}

func ReloadNginx() {

}

func getNginxContainerId(ctx context.Context, cli *client.Client) (containerId string, err error) {
	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: NginxContainerName}),
		All:     true,
	})
	if err != nil {
		return
	}
	if len(containers) == 0 {
		return "", nil
	}
	containerId = containers[0].ID
	return
}

func createNginxContainer(ctx context.Context, cli *client.Client) (containerId string, err error) {
	config := container.Config{
		Image: "nginx",
	}
	cwd, _ := os.Getwd()
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Target: "/etc/nginx/conf.d/default.conf",
				Source: filepath.Join(cwd, "config/nginx_default.conf"),
			},
		},
	}
	body, err := cli.ContainerCreate(ctx, &config, &hostConfig, nil, nil, NginxContainerName)
	if err != nil {
		return
	}
	containerId = body.ID
	return
}
