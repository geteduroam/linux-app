package network

import (
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network/inner"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network/method"
)

// A network belongs to the network interface when it has a method
type Network interface {
	// Method returns the EAP method
	Method() method.Type
	// ProviderInfo returns the EAP ProviderInfo
	ProviderInfo() ProviderInfo
}

// Help is the struct that contains information on how to contact an organization
type Help struct {
	// Email is the e-mail address as a string
	Email string
	// Phone is the phone number as a string
	Phone string
	// Web is the web URL as a string
	Web string
}

// ProviderInfo is the ProviderInfo element for the network
type ProviderInfo struct {
	// Helpdesk contains the help information on how to contact the organization that owns the network
	Helpdesk Help
	// Name is the display name of the network as provided by the organization which should be represented in the UI
	Name string
	// Description is the description of the network as provided by the organization
	Description string
	// Logo is the logo of the network, probably the logo of the organization in base64
	Logo string
	// Terms is the terms of use for this network
	Terms string
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
	// ProviderInfo is the ProviderInfo info
	ProviderInfo ProviderInfo
}

// Credentials is the credentials belonging to the Non TLS network
type Credentials struct {
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
}

// NonTLS is a structure for creating a network that has EAP method not TLS
type NonTLS struct {
	Base
	// Credentials are the credentials belonging to the Non TLS network
	Credentials Credentials
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

func (n *NonTLS) ProviderInfo() ProviderInfo {
	return n.Base.ProviderInfo
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

func (t *TLS) ProviderInfo() ProviderInfo {
	return t.Base.ProviderInfo
}
