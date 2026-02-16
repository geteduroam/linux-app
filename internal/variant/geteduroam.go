//go:build !getgovroam

// Package variant implements settings for different geteduroam variants
package variant

const (
	// AppID is the application ID for geteduroam
	AppID string = "app.eduroam.geteduroam"
	// DiscoveryURL is the discovery URL for geteduroam
	DiscoveryURL string = "https://discovery.eduroam.app/v3/discovery.json"
	// DisplayName is the display name for geteduroam
	DisplayName string = "geteduroam"
	// ProfileName is the connection profile name for geteduroam
	ProfileName string = "eduroam"
)
