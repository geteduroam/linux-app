package eap

// InnerAuthType defines the inner authentication methods that are returned by the EAP xml
type InnerAuthType int

const (
	NONE              InnerAuthType = 0
	PAP                             = 1
	MSCHAP                          = 2
	MSCHAPV2                        = 3
	EAP_PEAP_MSCHAPV2               = 25
	EAP_MSCHAPV2                    = 26
)

// ValidInnerAuth returns whether or not an integer is a valid inner authentication type
func ValidInnerAuth(input int) bool {
	switch InnerAuthType(input) {
	case NONE:
		return true
	case PAP:
		return true
	case MSCHAP:
		return true
	case MSCHAPV2:
		return true
	case EAP_PEAP_MSCHAPV2:
		return true
	case EAP_MSCHAPV2:
		return true
	}
	return false
}
