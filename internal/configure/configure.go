// package configure configures the eduroam connection by parsing the byte array
// It has handlers for UI events
package configure

import (
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/eap"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/nm"
)

// Configure is the structure that holds the handlers for UI events
// Handlers are just functions that are called to get certain data
type Config struct {
	// UsernameH is the handler for asking for the username
	// p is the prefix for the username that must be prefilled
	// s is the suffix for the username that must be prefilled
	UsernameH func(p string, s string) string

	// PasswordH is the handler for asking for the password
	// p is the prefix for the username that must be prefilled
	// s is the suffix for the username that must be prefilled
	PasswordH func() string

	// CertificateH is the handler for asking for the client certificate from the user
	// The handler is responsible for decrypting the certificate
	CertificateH func(name string, desc string) string

	// ProviderInfo.DisplayName from EAP metadata
	Displayname string

	// ProviderInfo.Description from EAP metadata
	Description string

	// The parsed network configuration
	Network network.Network
}

// Parse parses the connection using the EAP byte array
// It parses the config
func (c Config) Parse(config []byte) (Config, error) {
	// First we parse the config
	unpack, err := eap.Parse(config)
	if err != nil {
		return c, err
	}

	if len(unpack.EAPIdentityProvider.ProviderInfo.DisplayName) >= 1 {
		c.Displayname = unpack.EAPIdentityProvider.ProviderInfo.DisplayName[0].Value
	} else {
		c.Displayname = "unknown"
	}
	if len(unpack.EAPIdentityProvider.ProviderInfo.Description) >= 1 {
		c.Description = unpack.EAPIdentityProvider.ProviderInfo.Description[0].Value
	} else {
		c.Description = "unknown"
	}

	n, err := unpack.Network()
	if err != nil {
		return c, err
	}

	c.Network = n

	return c, nil
}

// Configure configures the connection using the parsed configuration
// It installs it using NetworkManager
func (c Config) Configure() error {
	switch t := c.Network.(type) {
	case *network.NonTLS:
		username := t.Username
		password := t.Password
		if username == "" {
			username = c.UsernameH(t.Prefix, t.Suffix)
		}
		if password == "" {
			password = c.PasswordH()
		}
		t.Username = username
		t.Password = password
		nm.Install(*t)
	default:
		panic("TLS networks are not yet supported")
	}

	return nil
}
