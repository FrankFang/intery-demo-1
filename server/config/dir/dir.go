package dir

import (
	"os"
	"path/filepath"
	"strconv"
)

func EnsureUserDir(userId uint) (userDir string) {
	cwd, _ := os.Getwd()
	userDir = filepath.Join(cwd, "/userspace/", strconv.Itoa(int(userId)))
	os.MkdirAll(userDir, os.ModePerm)
	return
}

func EnsureProjectDir(userDir string, projectId uint) (projectDir string) {
	projectDir = filepath.Join(userDir, strconv.Itoa(int(projectId)))
	os.MkdirAll(projectDir, os.ModePerm)
	return
}

func GetNginxConfigDir() (nginxConfigPath string) {
	nginxConfigPath = os.Getenv("NGINX_CONFIG_DIR")
	if nginxConfigPath == "" {
		cwd, _ := os.Getwd()
		nginxConfigPath = filepath.Join(cwd, "config", "nginx")
	}
	return
}

func GetSocketDir() (socketDir string) {
	socketDir = os.Getenv("SOCKET_DIR")
	if socketDir == "" {
		cwd, _ := os.Getwd()
		socketDir = filepath.Join(cwd, "userspace", "socket")
	}
	return
}
