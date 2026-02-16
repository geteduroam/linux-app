package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os/exec"
	"time"

	"codeberg.org/jwijenbergh/eduoauth-go/v2"
)

// Profile is the profile from discovery
type Profile struct {
	ID                   string           `json:"id"`
	EapConfigEndpoint    string           `json:"eapconfig_endpoint"`
	MobileConfigEndpoint string           `json:"mobileconfig_endpoint"`
	LetsWifiEndpoint     string           `json:"letswifi_endpoint"`
	WebviewEndpoint      string           `json:"webview_endpoint"`
	Name                 LocalizedStrings `json:"name"`
	Type                 string           `json:"type"`
	CachedResponse       []byte           `json:"-"`
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
	switch p.Type {
	case "webview":
		return RedirectFlow
	case "eap-config":
		return DirectFlow
	default:
		return OAuthFlow
	}
}

// RedirectURI gets the redirect URI from the profile
// It does some additional work by:
// - Checking if the redirect URI is a URL
// - Setting the scheme to HTTPS
func (p *Profile) RedirectURI() (string, error) {
	if p.WebviewEndpoint == "" {
		return "", errors.New("no redirect found")
	}
	u, err := url.Parse(p.WebviewEndpoint)
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
	defer res.Body.Close() //nolint:errcheck
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
	if p.CachedResponse != nil {
		return p.CachedResponse, nil
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
	return readResponse(res)
}

type letsWifiEndpoints struct {
	Href string `json:"href"`
	API  struct {
		AuthorizationEndpoint string `json:"authorization_endpoint"`
		TokenEndpoint         string `json:"token_endpoint"`
		EapConfigEndpoint     string `json:"eapconfig_endpoint"`
		MobileConfigEndpoint  string `json:"mobileconfig_endpoint"`
	} `json:"http://letswifi.app/api#v2"`
}

func (p *Profile) getLetsWifiEndpoints() ([]byte, error) {
	client := http.Client{Timeout: 10 * time.Second}
	if p.LetsWifiEndpoint == "" {
		return nil, errors.New("no Let's Wifi endpoint found")
	}
	req, err := http.NewRequest("GET", p.LetsWifiEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	b, err := readResponse(res)
	if err != nil {
		return nil, err
	}
	return b, nil
}

// EAPOAuth gets the EAP metadata using OAuth
func (p *Profile) EAPOAuth(ctx context.Context, auth func(authURL string)) ([]byte, error) {
	var err error
	b := p.CachedResponse
	if b == nil {
		b, err = p.getLetsWifiEndpoints()
		if err != nil {
			return nil, err
		}
	}
	var ep letsWifiEndpoints
	err = json.Unmarshal(b, &ep)
	if err != nil {
		return nil, err
	}

	o := eduoauth.OAuth{
		ClientID: "app.geteduroam.sh",
		EndpointFunc: func(context.Context) (*eduoauth.EndpointResponse, error) {
			return &eduoauth.EndpointResponse{
				AuthorizationURL: ep.API.AuthorizationEndpoint,
				TokenURL:         ep.API.TokenEndpoint,
			}, nil
		},
		RedirectPath: "/",
	}
	url, err := o.AuthURL(ctx, "eap-metadata")
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
	err = o.Exchange(ctx, "")
	if err != nil {
		return nil, err
	}

	c := o.NewHTTPClient()
	req, err := http.NewRequestWithContext(ctx, "POST", ep.API.EapConfigEndpoint, nil)
	if err != nil {
		return nil, err
	}
	res, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	return readResponse(res)
}
