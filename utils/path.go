package utils

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)


func PathExist(_path string) bool {
	_, err := os.Stat(_path)
	if err != nil && os.IsNotExist(err) {
		return false
	}
	return true
}

var exeDir string

func GetExeDir() string {
	if exeDir != "" {
		return exeDir
	}
	file, _ := exec.LookPath(os.Args[0])
	path, _ := filepath.Abs(file)
	exeDir = filepath.Dir(path)

	return exeDir
}

func IsRelactivePath(path string) bool {
	if strings.Index(path, ".") == 0 {
		return true
	}
	return false
}

func AbsPath(path string) string {
	if IsRelactivePath(path) {
		path = GetExeDir() + string(os.PathSeparator) + path
		return path
	}
	return path
}


