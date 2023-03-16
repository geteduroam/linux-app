// package config has methods to write (TODO: read) config files
package config

import (
	"os"
	"path"
)

// Config is the main structure for the configuration
// This right now is only the directory but in the future it would probably have more settings
type Config struct {
	Directory string
}

// ensureDir creates the directory and returns an error if it cannot be created
func ensureDir(dir string) error {
	// 700: read, write, execute only for the owner
	err := os.MkdirAll(dir, 0o700)
	if err != nil {
		return err
	}
	return nil
}

// Write writes the configuration to the filesystem with the filename and string
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
