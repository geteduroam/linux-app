package inner

import (
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network/method"
)

// Type defines the inner authentication methods that are returned by the EAP xml
type Type int

// TODO: Should we split these in EAP and non-EAP instead?
const (
	NONE     Type = 0
	PAP      Type = 1
	MSCHAP   Type = 2
	MSCHAPV2 Type = 3
	// TODO: remove this? https://github.com/geteduroam/windows-app/blob/f11f00dee3eb71abd38537e18881463f83b180d3/CHANGELOG.md?plain=1#L34
	EAP_PEAP_MSCHAPV2 Type = 25
	EAP_MSCHAPV2      Type = 26
)

// EAP returns whether the type is an EAP inner type
func (t Type) EAP() bool {
	switch t {
	case EAP_PEAP_MSCHAPV2:
		return true
	case EAP_MSCHAPV2:
		return true
	}
	return false
}

// String returns the string representation of the inner type
func (t Type) String() string {
	switch t {
	case PAP:
		return "pap"
	case MSCHAP:
		return "mschap"
	case MSCHAPV2:
		fallthrough
	case EAP_PEAP_MSCHAPV2:
		fallthrough
	case EAP_MSCHAPV2:
		return "mschapv2"
	}
	return ""
}

// Valid returns whether or not an integer is a valid inner authentication type
// See https://github.com/geteduroam/geteduroam-sh/blob/54044773812502487ad0f68898cd6b9e110cb0f6/eap-config.sh#L55
func Valid(mt method.Type, input int, eap bool) bool {
	// For TLS we do not have any inner, any is valid
	if mt == method.TLS {
		return true
	}
	// Check if the inner is an EAP or NON EAP type
	// They should match with what we expect it to be
	// So for example if we pass an input and expect an EAP type, but the input is actually NON-EAP, we return false as it's not valid
	if Type(input).EAP() != eap {
		return false
	}
	// For TTLS, we support PAP, MSCHAP, MSCHAPv2 and EAP MSCHAPV2
	if mt == method.TTLS {
		switch Type(input) {
		case PAP:
			return true
		case MSCHAP:
			return true
		case MSCHAPV2:
			return true
		case EAP_MSCHAPV2:
			return true
		}
		return false
	}
	// for PEAP, we only support EAP*MSCHAPV2
	if mt == method.PEAP {
		switch Type(input) {
		case EAP_PEAP_MSCHAPV2:
			return true
		case EAP_MSCHAPV2:
			return true
		}
		return false
	}
	return false
}
