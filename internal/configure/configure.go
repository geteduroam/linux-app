// package configure configures the eduroam connection by parsing the byte array
// It has handlers for UI events
package configure

import (
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/eap"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/nm"
)

type Configure struct {
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
}

func (c Configure) Configure(config []byte) error {
	// First we parse the config
	unpack, err := eap.Parse(config)
	if err != nil {
		return err
	}

	n, err := unpack.Network()
	if err != nil {
		return err
	}

	switch t := n.(type) {
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
