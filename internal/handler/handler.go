// package handlers handles the eduroam connection by parsing the byte array
// It has handlers for UI events
package handler

import (
	"time"

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/config"
	"github.com/geteduroam/linux-app/internal/eap"
	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/network/cert"
	"github.com/geteduroam/linux-app/internal/nm"
)

// Handlers is the structure that holds the handlers for UI events
// 'Handlers' are just functions that are called to get certain data
type Handlers struct {
	// CredentialsH is the handler for asking for the username and password
	// c are the credentials which also contains prefixes and suffixes for the username
	// pi is the provider info
	// It returns the username and password that were filled in
	CredentialsH func(c network.Credentials, pi network.ProviderInfo) (string, string, error)

	// CertificateH is the handler for asking for the client certificate from the user
	// It returns the certificate, the passphrase and an error
	CertificateH func(cert string, passphrase string, pi network.ProviderInfo) (string, string, error)
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
func (h Handlers) Configure(eap []byte) (*time.Time, error) {
	// Get the network
	n, err := h.network(eap)
	if err != nil {
		return nil, err
	}
	var uuid string

	// get the previous UUID if the config can be loaded
	c, err := config.Load()
	if err == nil {
		uuid = c.UUID
	}

	var valid *time.Time
	switch t := n.(type) {
	case *network.NonTLS:
		if t.Credentials.Username == "" || t.Credentials.Password == "" {
			username, password, cerr := h.CredentialsH(t.Credentials, n.ProviderInfo())
			if cerr != nil {
				slog.Debug("Error asking for credentials", "error", err)
				return nil, cerr
			}
			t.Credentials.Username = username
			t.Credentials.Password = password
		}
		uuid, err = nm.Install(*t, uuid)
	case *network.TLS:
		// if a PKCS12 file is uploaded by the user we expect it to be not base64 encoded
		b64 := t.RawPKCS12 != ""
		// TODO: Loop until the PKCS12 can be decrypted successfully?
		if t.ClientCert == nil {
			ccert, passphrase, err := h.CertificateH(t.RawPKCS12, t.Password, n.ProviderInfo())
			if err != nil {
				return nil, err
			}
			// here the data is not base64 encoded
			t.ClientCert, err = cert.NewClientCert(ccert, passphrase, b64)
			if err != nil {
				return nil, err
			}
		}
		v := t.Validity()
		valid = &v
		uuid, err = nm.InstallTLS(*t, uuid)
	default:
		panic("unsupported network")
	}
	if err != nil {
		slog.Debug("Error installing network", "error", err)
		return nil, err
	}
	// save the config with the uuid
	nc := config.Config{
		UUID:     uuid,
		Validity: valid,
	}
	err = nc.Write()
	if err != nil {
		slog.Debug("Error configuring network", "error", err)
		return nil, err
	}
	return valid, nil
}
