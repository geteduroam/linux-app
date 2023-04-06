package eap

import (
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network/inner"
	"testing"
	"os"
)

// These are very rudimentary tests, needs parametrization
func Test_Parse(t *testing.T) {
	var err error
	var b []byte
	var EIPL *EAPIdentityProviderList

	b, err = os.ReadFile("eduroam-eap-generic-eVAe.eap-config.xml")
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}
	EIPL, err = Parse(b)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	var EIP *EAPIdentityProvider
	EIP = EIPL.EAPIdentityProvider

	var methods *AuthenticationMethods
	methods, err = EIP.authenticationMethods()
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	var method *AuthenticationMethod

	// The first AuthenticationMethod
	method = methods.AuthenticationMethod[0]

	var PIAT inner.Type
	PIAT, err = method.preferredInnerAuthType(25)
	// t.Fatalf("preferred type: %d", PIAT)
	if PIAT.String() != "mschapv2" {
		t.Fatalf("preferred type: %s", PIAT)
	}
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	// The second AuthenticationMethod
	method = methods.AuthenticationMethod[1]
	PIAT, err = method.preferredInnerAuthType(25)
	if PIAT.String() != "mschapv2" {
		t.Fatalf("preferred type: %s", PIAT)
	}
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	// The third AuthenticationMethod, incompatible type
	method = methods.AuthenticationMethod[2]
	PIAT, err = method.preferredInnerAuthType(25)
	if PIAT.String() != "" {
		t.Fatalf("preferred type: %s", PIAT)
	}
	if err != nil && err.Error() != "no viable inner authentication method found" {
		t.Fatalf("Error: %s, %s", PIAT, err.Error())
	}
}
