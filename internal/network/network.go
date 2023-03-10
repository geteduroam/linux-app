package network

import (
	"github.com/jwijenbergh/geteduroam-linux/internal/network/inner"
	"github.com/jwijenbergh/geteduroam-linux/internal/network/method"
)

// A network belongs to the network interface when it has a method
type Network interface {
	// Method returns the EAP method
	Method() method.Type
}

// Base is the definition that each network always has
type Base struct {
	// Cert is the list of CA certificates that are used
	Cert []string
	// SSID is the name of the network
	SSID string
	// MinRSN is the minimum RSN proto
	MinRSN string
	// ServerIDs is the list of server names
	ServerIDs []string
}

// NonTLS is a structure for creating a network that has EAP method not TLS
type NonTLS struct {
	Base
	// Username is the string that is configured as the identity for the connection
	// This is gotten from the user
	// This username is prefixed with the InnerIdentityPrefix
	// It is suffixed with the InnerIdentitySuffix
	Username string
	// Prefix is the prefix for the username
	Prefix string
	// Suffix is the suffix for the username
	Suffix string
	// Password is the string that is configured as the RADIUS password for the connection
	Password string
	// MethodType is the method
	MethodType method.Type
	// InnerAuth is the inner authentication method
	InnerAuth inner.Type
	// AnonIdentity is the anonymous identity found in the EAP config, OuterIdentity in clientcredentials
	// This is optional as when it's not set it will be set to the username
	AnonIdentity string
}

func (n *NonTLS) Method() method.Type {
	return n.MethodType
}

// TLS is a structure for creating a network that has EAP method TLS
type TLS struct {
	Base
	// ClientCertificate is the client certificate that is optionally protected by a password
	ClientCertificate string

	// Password is the password that encrypts the ClientCertificate
	Password string
}

func (t *TLS) Method() method.Type {
	return method.TLS
}
