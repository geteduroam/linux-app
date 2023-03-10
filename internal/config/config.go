package config

import (
	"os"
	"path"
)

type Config struct {
	Directory string
}

func ensureDir(dir string) error {
	// 700: read, write, execute only for the owner
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return err
	}
	return nil
}

func (c *Config) Write(filename string, content string) (string, error) {
	if err := ensureDir(c.Directory); err != nil {
		return "", err
	}
	fpath := path.Join(c.Directory, filename)
	if err := os.WriteFile(fpath, []byte(content), 0o600); err != nil {
		return "", err
	}
	return fpath, nil
}
