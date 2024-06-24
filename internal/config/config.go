// package config has methods to write (TODO: read) config files
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/utils"
)

// Config is the main structure for the configuration
type Config struct {
	UUID     string     `json:"uuid"`
	Validity *time.Time `json:"validity,omitempty"`
}

// Versioned contains the actual config data prefixed with a version field when marshalled as JSON
type Versioned struct {
	// Config is the versioned configuration
	// It is versioned so that we can change the version and migrate older configs in the future
	Config Config `json:"v1"`
}

// Directory returns the directory where the config files are stored
func Directory() (p string, err error) {
	// This follows the XDG specification at https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
	//  From that doc: $XDG_DATA_HOME defines the base directory relative to which user-specific data files should be stored. If $XDG_DATA_HOME is either not set or empty, a default equal to $HOME/.local/share should be used.
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			slog.Debug("Error finding user HomeDir", "error", err)
			return "", err
		}
		dir = filepath.Join(home, ".local/share")
	}
	p = filepath.Join(dir, "geteduroam")
	return
}

// Write writes the configuration to the filesystem with the filename and string
func WriteFile(filename string, content []byte) (string, error) {
	dir, err := Directory()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	fpath := filepath.Join(dir, filename)
	if err := os.WriteFile(fpath, content, 0o600); err != nil {
		return "", err
	}
	return fpath, nil
}

var configName = "state"

// Load loads the configuration from the state
func Load() (*Config, error) {
	dir, err := Directory()
	if err != nil {
		return nil, err
	}

	p := filepath.Join(dir, configName)

	b, err := os.ReadFile(p)
	if err != nil {
		slog.Debug("Error reading config file", "file", p, "error", err)
		return nil, err
	}

	var v Versioned

	err = json.Unmarshal(b, &v)
	if err != nil {
		slog.Debug("Error reading config file", "file", p, "error", err)
		return nil, err
	}
	utils.Verbosef("Reading config file %s", p)
	return &v.Config, nil
}

// Write writes the configuration to the state
func (c Config) Write() (err error) {
	// we pack the struct in a versioned struct
	// This is so that we can in the future migrate configs if we drastically change the format
	// marshal the config
	v := &Versioned{
		Config: c,
	}
	b, err := json.Marshal(&v)
	if err != nil {
		slog.Debug("Error writing config file", "file", configName, "error", err)
		return err
	}
	_, err = WriteFile(configName, b)
	if err != nil {
		slog.Debug("Error writing config file", "file", configName, "error", err)
	}
	return
}
