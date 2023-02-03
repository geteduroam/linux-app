package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/jwijenbergh/geteduroam-linux/internal/instance"
)

type Discovery struct {
	Instances []instance.Instance `json:"instances"`
	// See: https://github.com/geteduroam/windows-app/blob/22cd90f36031907c7174fbdc678edafaa627ce49/CHANGELOG.md#changed
	Seq     int `json:"seq"`
	Version int `json:"version"`
}

type Cache struct {
	// Cached is the cached list of discovery
	Cached Discovery `json:"previous"`
	// Seq is the parsed sequence number
	Seq Seq `json:"seq"`
	// LastUpdate is the last time we updated the cache
	LastUpdate time.Time `json:"updated"`
}

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
	return !n.After(u)
}

// Instances gets the instances either from the cache or from scratch
func (c *Cache) Instances() (*[]instance.Instance, error) {
	if !c.ToUpdate() {
		return &c.Cached.Instances, nil
	}

	req, err := http.NewRequest("GET", "https://discovery.eduroam.app/v1/discovery.json", nil)
	if err != nil {
		return &c.Cached.Instances, err
	}

	// Do request
	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return &c.Cached.Instances, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return &c.Cached.Instances, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return &c.Cached.Instances, fmt.Errorf("status code is not 2xx for discovery. Status code: %v, body: %v", res.StatusCode, string(body))
	}

	var d *Discovery
	err = json.Unmarshal(body, &d)
	if err != nil {
		return &c.Cached.Instances, err
	}

	v, err := NewSeq(d.Seq)
	if err != nil {
		return &c.Cached.Instances, err
	}
	// Do not accept older versions
	// This happens if the cached version is newer
	if c.Seq.After(*v) {
		return &c.Cached.Instances, fmt.Errorf("cached seq is newer")
	}
	c.Cached = *d
	c.Seq = *v
	c.LastUpdate = time.Now()
	return &d.Instances, nil
}
