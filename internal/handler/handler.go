// package handlers handles the eduroam connection by parsing the byte array
// It has handlers for UI events
package handler

import (
	"github.com/geteduroam/linux-app/internal/eap"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/nm"
)

// Handlers is the structure that holds the handlers for UI events
// 'Handlers' are just functions that are called to get certain data
type Handlers struct {
	// CredentialsH is the handler for asking for the username and password
	// c are the credentials which also contains prefixes and suffixes for the username
	// pi is the provider info
	// It returns the username and password that were filled in
	CredentialsH func(c network.Credentials, pi network.ProviderInfo) (string, string)

	// CertificateH is the handler for asking for the client certificate from the user
	// The handler is responsible for decrypting the certificate
	CertificateH func(cert string, pi network.ProviderInfo) string
}

// network gets the network by parsing the connection using the EAP byte array
func (h Handlers) network(config []byte) (network.Network, error) {
	// First we parse the config
	unpack, err := eap.Parse(config)
	if err != nil {
		return nil, err
	}

	n, err := unpack.Network()
	if err != nil {
		return nil, err
	}

	return n, nil
}

// Configure configures the connection using the parsed configuration
// It installs it using NetworkManager
func (h Handlers) Configure(config []byte) error {
	// Get the network
	n, err := h.network(config)
	if err != nil {
		return err
	}
	switch t := n.(type) {
	case *network.NonTLS:
		username, password := h.CredentialsH(t.Credentials, n.ProviderInfo())
		t.Credentials.Username = username
		t.Credentials.Password = password
		return nm.Install(*t)
	default:
		panic("TLS networks are not yet supported")
	}
}
