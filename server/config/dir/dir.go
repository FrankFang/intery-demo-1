package dir

import (
	"os"
	"path/filepath"
	"strconv"
)

func EnsureUserDir(userId uint) (userDir string) {
	cwd, _ := os.Getwd()
	userDir = filepath.Join(cwd, "/userspace/", "user"+strconv.Itoa(int(userId)))
	os.MkdirAll(userDir, os.ModePerm)
	return
}

func EnsureProjectDir(userDir string, projectId uint) (projectDir string) {
	projectDir = filepath.Join(userDir, "project"+strconv.Itoa(int(projectId)))
	os.MkdirAll(projectDir, os.ModePerm)
	return
}

func ResetDir(anyDir string) {
	os.RemoveAll(anyDir)
	os.MkdirAll(anyDir, os.ModePerm)
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

func GetAppTemplatesDir(appKind string) (appTemplatesDir string) {
	appTemplatesDir = os.Getenv("APP_TEMPLATES_DIR")
	if appTemplatesDir == "" {
		cwd, _ := os.Getwd()
		appTemplatesDir = filepath.Join(cwd, "server/app-templates")
	}
	appTemplatesDir = filepath.Join(appTemplatesDir, appKind)
	return
}

func GetLogDir() (logDir string) {
	logDir = os.Getenv("LOG_DIR")
	if logDir == "" {
		cwd, _ := os.Getwd()
		logDir = filepath.Join(cwd, "log")
	}
	return
}

func GetKeyDir() (keyDir string) {
	keyDir = os.Getenv("KEY_DIR")
	if keyDir == "" {
		cwd, _ := os.Getwd()
		keyDir = filepath.Join(cwd, "key")
	}
	return
}

func GetFrontendDir() (frontendDir string) {
	frontendDir = os.Getenv("FRONTEND_DIR")
	if frontendDir == "" {
		cwd, _ := os.Getwd()
		frontendDir = filepath.Join(cwd, "www")
	}
	return
}
