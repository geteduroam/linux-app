package eap

// MethodType defines the EAP methods that are returned by the EAP xml
type MethodType int

const (
	TLS  MethodType = 13
	TTLS            = 21
	PEAP            = 25
)

// ValidMethod returns whether or not an integer is a valid method type
func ValidMethod(input int) bool {
	switch MethodType(input) {
	case TLS:
		return true
	case TTLS:
		return true
	case PEAP:
		return true
	}
	return false
}

// NeedsCredentials returns whether or not this EAP method needs credentials (username/password) from the user
func (m MethodType) NeedsCredentials() bool {
	return m != TLS
}

// NeedsCertificate returns whether or not this EAP method needs a client certificate
// TODO: have a separate method that reports if the certificate needs a password
func (m MethodType) NeedsCertificate() bool {
	return m == TLS
}