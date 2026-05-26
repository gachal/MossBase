package config

import (
	"os"
	"path/filepath"
)

const sentinelFile = ".installed"

func configDir() string {
	return "configs"
}

func IsInstalled() bool {
	_, err := os.Stat(filepath.Join(configDir(), sentinelFile))
	return err == nil
}

func MarkInstalled() error {
	if err := os.MkdirAll(configDir(), 0o755); err != nil {
		return err
	}
	f, err := os.Create(filepath.Join(configDir(), sentinelFile))
	if err != nil {
		return err
	}
	f.Close()
	return nil
}
