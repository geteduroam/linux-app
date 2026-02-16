// Package method implements EAP methods
package method

// Type defines the EAP methods that are returned by the EAP xml
type Type int

const (
	// TLS is the TLS EAP Method
	TLS Type = 13
	// TTLS is the TTLS EAP Method
	TTLS Type = 21
	// PEAP is the PEAP EAP Method
	PEAP Type = 25
)

// IsValid returns whether or not an integer is a valid method type
func IsValid(input int) bool {
	switch Type(input) {
	case TLS:
		return true
	case TTLS:
		return true
	case PEAP:
		return true
	}
	return false
}

func (m Type) String() string {
	switch m {
	case TLS:
		return "tls"
	case TTLS:
		return "ttls"
	case PEAP:
		return "peap"
	}
	return ""
}

// NeedsCredentials returns whether or not this method needs credentials (username/password) from the user
func (m Type) NeedsCredentials() bool {
	return m != TLS
}

// NeedsCertificate returns whether or not this EAP method needs a client certificate
// TODO: have a separate method that reports if the certificate needs a password
func (m Type) NeedsCertificate() bool {
	return m == TLS
}
