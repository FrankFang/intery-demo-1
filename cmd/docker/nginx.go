package docker

import (
	"context"
	"errors"
	"os"
	"path/filepath"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const NginxContainerName = "nginx1"

func StartNginxContainer(ctx context.Context) (containerId string, err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}
	containerId, err = getNginxContainerId(ctx, cli)
	if err != nil {
		return
	}
	if containerId == "" {
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
		Image:        "nginx",
		ExposedPorts: nat.PortSet{"80": struct{}{}},
	}
	cwd, _ := os.Getwd()
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Target: "/etc/nginx/conf.d/default.conf",
				Source: filepath.Join(cwd, "config/nginx_default.conf"),
			},
			{
				Type:   "bind",
				Target: "/tmp/socket",
				Source: filepath.Join(cwd, "userspace/socket/"),
			},
		},
		PortBindings: nat.PortMap{
			"80/tcp": []nat.PortBinding{
				{
					HostIP:   "",
					HostPort: "80",
				},
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

func ReloadNginx(ctx context.Context) (err error) {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return
	}
	containerId, err := getNginxContainerId(ctx, cli)
	if err != nil {
		return
	}
	if containerId == "" {
		containerId, err = StartNginxContainer(ctx)
		if err != nil {
			return errors.New("start nginx container failed")
		}
	}
	inspect, err := cli.ContainerInspect(ctx, containerId)
	if err != nil {
		return
	}
	if !inspect.State.Running {
		err = cli.ContainerStart(ctx, containerId, types.ContainerStartOptions{})
		if err != nil {
			return
		}
	}
	exec, err := cli.ContainerExecCreate(ctx, containerId, types.ExecConfig{
		Cmd: []string{"nginx", "-s", "reload"},
	})
	if err != nil {
		return
	}
	err = cli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return
	}
	exec, err = cli.ContainerExecCreate(ctx, containerId, types.ExecConfig{
		Cmd: []string{"chmod", "777", "-R", "/tmp/socket"},
	})
	if err != nil {
		return
	}
	err = cli.ContainerExecStart(ctx, exec.ID, types.ExecStartCheck{})
	if err != nil {
		return
	}
	return
}
