package instance

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"github.com/jwijenbergh/eduoauth-go"
)

// Profile is the profile from discovery
type Profile struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	Default               bool   `json:"default"`
	EapConfigEndpoint     string `json:"eapconfig_endpoint"`
	ID                    string `json:"id"`
	Name                  string `json:"name"`
	OAuth                 bool   `json:"oauth"`
	TokenEndpoint         string `json:"token_endpoint"`
	Redirect              string `json:"redirect"`
}

// FlowCode is the type of flow that we will use to get the EAP config
type FlowCode int8

const (
	// DirectFlow tells us that we can get the EAP config directly without OAuth
	DirectFlow FlowCode = iota
	// RedirectFlow tells us we need to follow the redirect
	RedirectFlow
	// OAuthFlow tells us we can get the EAP config through OAuth
	OAuthFlow
)

// Flow gets the flow we need to go through to get the EAP config
// See: https://github.com/geteduroam/cattenbak/blob/481e243f22b40e1d8d48ecac2b85705b8cb48494/cattenbak.py#L68
func (p *Profile) Flow() FlowCode {
	// A Redirect entry is present
	// This means that we need to follow the URI in the redirect flow
	if p.Redirect != "" {
		return RedirectFlow
	}
	// OAuth is present, we need to get the EAP through some OAuth flow
	if p.OAuth {
		return OAuthFlow
	}
	// Get the config directly
	return DirectFlow
}

// RedirectURI gets the redirect URI from the profile
// It does some additional work by:
// - Checking if the redirect URI is a URL
// - Setting the scheme to HTTPS
func (p *Profile) RedirectURI() (string, error) {
	if p.Redirect == "" {
		return "", errors.New("no redirect found")
	}
	u, err := url.Parse(p.Redirect)
	if err != nil {
		return "", err
	}
	// We enforce HTTPS
	if u.Scheme != "https" {
		u.Scheme = "https"
	}
	return u.String(), nil
}

// readResponse reads the HTTP response and returns the body and error
// It also ensures the body is closed in the end to prevent a resource leak
func readResponse(res *http.Response) ([]byte, error) {
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

// EAPDirect Gets an EAP config using the direct flow
// It returns the byte array of the EAP config and an error if there is one
func (p *Profile) EAPDirect() ([]byte, error) {
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
	return readResponse(res)
}

// EAPOAuth gets the EAP metadata using OAuth
func (p *Profile) EAPOAuth(auth func(authURL string)) ([]byte, error) {
	o := eduoauth.OAuth{
		ClientID:             "app.geteduroam.sh",
		BaseAuthorizationURL: p.AuthorizationEndpoint,
		TokenURL:             p.TokenEndpoint,
		RedirectPath:         "/",
	}
	url, err := o.AuthURL("eap-metadata")
	if err != nil {
		return nil, err
	}

	// Open the authorization screen in a goroutine
	// TODO: make this return an error and use channels to communicate it?
	go auth(url)
	err = exec.Command("xdg-open", url).Start()
	if err != nil {
		return nil, err
	}
	err = o.Exchange(context.Background(), "")
	if err != nil {
		return nil, err
	}

	c := o.NewHTTPClient()
	req, err := http.NewRequestWithContext(context.Background(), "POST", p.EapConfigEndpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return readResponse(res)
}
