// Package discovery contains methods to parse the discovery format from https://discovery.eduroam.app/v3/discovery.json into providers
package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/provider"
	"github.com/geteduroam/linux-app/internal/variant"
)

// Discovery is the main structure that is used for unmarshalling the JSON
type Discovery struct {
	Value Value `json:"http://letswifi.app/discovery#v3"`
}

// Value is the discovery JSON value
type Value struct {
	Providers provider.Providers `json:"providers"`
	// See: https://github.com/geteduroam/windows-app/blob/22cd90f36031907c7174fbdc678edafaa627ce49/CHANGELOG.md#changed
	Seq int `json:"seq"`
}

// Cache is the cached discovery list
// TODO: This should be read from disk so that the app can function when the discovery is offline
type Cache struct {
	// Cached is the cached list of discovery
	Cached Discovery `json:"previous"`
	// LastUpdate is the last time we updated the cache
	LastUpdate time.Time `json:"updated"`
}

// NewCache creates a new cache struct
func NewCache() *Cache {
	return &Cache{}
}

// ToUpdate returns whether or not we should update the cached list
func (c *Cache) ToUpdate() bool {
	if c.LastUpdate.IsZero() {
		return true
	}
	// We update every hour
	u := c.LastUpdate.Add(1 * time.Hour)
	n := time.Now()
	return n.After(u)
}

// Providers gets the providers either from the cache or from scratch
func (c *Cache) Providers() (*provider.Providers, error) {
	if !c.ToUpdate() {
		return &c.Cached.Value.Providers, nil
	}

	req, err := http.NewRequest("GET", variant.DiscoveryURL, nil)
	if err != nil {
		return &c.Cached.Value.Providers, err
	}

	// Do request
	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		slog.Debug("Error requesting discovery.json", "error", err)
		return &c.Cached.Value.Providers, err
	}
	defer res.Body.Close() //nolint:errcheck

	body, err := io.ReadAll(res.Body)
	if err != nil {
		slog.Debug("Error reading discovery.json response", "error", err)
		return &c.Cached.Value.Providers, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return &c.Cached.Value.Providers, fmt.Errorf("status code is not 2xx for discovery. Status code: %v, body: %v", res.StatusCode, string(body))
	}

	var d *Discovery
	err = json.Unmarshal(body, &d)
	if err != nil {
		slog.Debug("Error loading discovery.json", "error", err)
		return &c.Cached.Value.Providers, err
	}

	// Do not accept older versions
	// This happens if the cached version is higher
	if c.Cached.Value.Seq > d.Value.Seq {
		return &c.Cached.Value.Providers, fmt.Errorf("cached seq is higher")
	}

	c.Cached = *d
	c.LastUpdate = time.Now()
	return &d.Value.Providers, nil
}
