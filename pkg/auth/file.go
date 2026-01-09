package auth

import (
	"os"
	"path/filepath"
)

const (
	rootDirName  = ".yt-analytics-mcp"
	dirWritePerm = 0755
)

func createAppDir(appRootPath string) error {
	if _, err := os.Stat(appRootPath); os.IsNotExist(err) {
		return os.MkdirAll(appRootPath, os.FileMode(dirWritePerm))
	}
	return nil
}

func GetAppRootDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	appRootPath := filepath.Join(home, rootDirName)

	if err = createAppDir(appRootPath); err != nil {
		return "", err
	}

	return appRootPath, nil
}
