// package discovery contains methods to parse the discovery format from https://discovery.eduroam.app/v1/discovery.json into instances
package discovery

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/geteduroam/linux-app/internal/instance"
)

// Discovery is the main structure that is used for unmarshalling the JSON
type Discovery struct {
	Instances instance.Instances `json:"instances"`
	// See: https://github.com/geteduroam/windows-app/blob/22cd90f36031907c7174fbdc678edafaa627ce49/CHANGELOG.md#changed
	Seq     int `json:"seq"`
	Version int `json:"version"`
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
	return !n.After(u)
}

// Instances gets the instances either from the cache or from scratch
func (c *Cache) Instances() (*instance.Instances, error) {
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

	d.Instances = append(d.Instances, instance.Instance{
		Name: "LetsWifi development banaan",
		Profiles: []instance.Profile{
			{
				AuthorizationEndpoint: "http://0.0.0.0:8080/oauth/authorize/",
				Default:               true,
				EapConfigEndpoint:     "http://0.0.0.0:8080/api/eap-config/",
				OAuth:                 true,
				TokenEndpoint:         "http://0.0.0.0:8080/oauth/token/",
			},
		},
	})

	// Do not accept older versions
	// This happens if the cached version is higher
	if c.Cached.Seq > d.Seq {
		return &c.Cached.Instances, fmt.Errorf("cached seq is higher")
	}

	c.Cached = *d
	c.LastUpdate = time.Now()
	return &d.Instances, nil
}
