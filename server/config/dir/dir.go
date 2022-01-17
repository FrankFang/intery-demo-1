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
