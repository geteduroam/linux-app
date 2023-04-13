package eap

import (
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network/inner"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/network/method"
	"gitlab.geant.org/TI_Incubator/geteduroam-linux/internal/utils"
	"testing"
	"os"
)

// These are very rudimentary tests, needs parametrization
func Test_Parse(t *testing.T) {
	var err error
	var b []byte

	b, err = os.ReadFile("eduroam-eap-generic-eVAe.eap-config.xml")
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	var EIPL *EAPIdentityProviderList
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

	cases := []struct {
		authmethod    int
		preferred     method.Type
		want          string
		err           string
	}{
		{
			// The first AuthenticationMethod
			authmethod: 0,
			preferred:  25,
			want:       "mschapv2",
			err:        "",
		},
		{
			// The second AuthenticationMethod
			authmethod: 1,
			preferred:  25,
			want:       "mschapv2",
			err:        "",
		},
		{
			// The third AuthenticationMethod
			authmethod: 2,
			preferred:  25,
			want:       "",
			err:        "no viable inner authentication method found",
		},
	}

	var m *AuthenticationMethod
	var r inner.Type

	for _, c := range cases {
		m = methods.AuthenticationMethod[c.authmethod]
		r, err = m.preferredInnerAuthType(c.preferred)
		es := utils.EtoString(err)
		es := utils.ErrorString(err)
		if r.String() != c.want || es != c.err {
			t.Fatalf("Result: %s, %s Want: %s, %s", r, es, c.want, c.err)
		}
	}
}
