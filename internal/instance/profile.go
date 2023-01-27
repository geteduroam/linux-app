package instance

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

type Profile struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	Default               bool   `json:"default"`
	EapConfigEndpoint     string `json:"eapconfig_endpoint"`
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	OAuth                 bool   `json:"oauth"`
	TokenEndpoint         string `json:"string"`
	Redirect              string `json:"redirect"`
}


func (p *Profile) EAP() ([]byte, error) {
	if p.OAuth {
		panic("no oauth support just yet")
	}

	// Do request
	req, err := http.NewRequest("GET", p.EapConfigEndpoint, nil)
	if err != nil {
		return nil, err
	}

	client := http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode > 299 {
		return nil, fmt.Errorf("status code is not 2xx for eap. Status code: %v, body: %v", res.StatusCode, string(body))
	}
	return body, nil
}
