package docker

import (
	"context"
	"errors"
	"intery/server/config/dir"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

const NginxContainerName = "nginx1"

var cli *client.Client = nil

func init() {
	c, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatal("Create client failed. ", err)
	}
	cli = c
}

func StartNginxContainer(c context.Context) (containerId string, err error) {
	containerId, err = GetNginxContainerId(c)
	if err != nil {
		log.Fatal("Get container id failed. ", err)
	}
	if containerId == "" {
		containerId, err = CreateNginxContainer(c)
		if err != nil {
			log.Fatal("Create container failed. ", err)
		}
	}

	inspect, _ := cli.ContainerInspect(c, containerId)
	if inspect.State.Running {
		return
	}
	err = cli.ContainerStart(c, containerId, types.ContainerStartOptions{})
	if err != nil {
		log.Fatal("Start container failed. ", err)
	}
	exec, err := cli.ContainerExecCreate(c, containerId, types.ExecConfig{
		Cmd: []string{"/bin/sh", "-c", "while sleep 3; do chmod 777 /tmp/socket/*.sock; done"},
	})
	if err != nil {
		log.Fatal("Create exec failed. ", err)
	}
	err = cli.ContainerExecStart(c, exec.ID, types.ExecStartCheck{})
	if err != nil {
		log.Fatal("Start exec failed. ", err)
	}
	return
}

func GetNginxContainerId(ctx context.Context) (containerId string, err error) {
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

func CreateNginxContainer(c context.Context) (containerId string, err error) {
	config := container.Config{
		Image:        "nginx",
		ExposedPorts: nat.PortSet{"80": struct{}{}},
	}
	hostConfig := container.HostConfig{
		Mounts: []mount.Mount{
			{
				Type:   "bind",
				Target: "/etc/nginx/conf.d",
				Source: dir.GetNginxConfigDir(),
			},
			{
				Type:   "bind",
				Target: "/tmp/socket",
				Source: dir.GetSocketDir(),
			},
			{
				Type:   "bind",
				Target: "/tmp/log",
				Source: dir.GetLogDir(),
			},
			{
				Type:   "bind",
				Target: "/key",
				Source: dir.GetKeyDir(),
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
	body, err := cli.ContainerCreate(c, &config, &hostConfig, nil, nil, NginxContainerName)
	if err != nil {
		log.Println("Create container failed. ", err)
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
	containerId, err := GetNginxContainerId(ctx)
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

func RemoveContainer(ctx context.Context, containerId string) (err error) {
	err = cli.ContainerRemove(ctx, containerId, types.ContainerRemoveOptions{
		RemoveVolumes: false,
		Force:         true,
	})
	if err != nil {
		return
	}
	return
}
