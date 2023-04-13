package eap

import (
	"github.com/geteduroam/linux-app/internal/network/method"
	"github.com/geteduroam/linux-app/internal/utils"
	"testing"
	"os"
)

// These are very rudimentary tests, needs parametrization
func Test_Parse(t *testing.T) {
	var err error
	var b []byte

	b, err = os.ReadFile("test_data/eduroam-eap-generic-eVAe.eap-config.xml")
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	EIPL, err := Parse(b)
	if err != nil {
		t.Fatalf("Error: %s", err.Error())
	}

	EIP := EIPL.EAPIdentityProvider

	methods, err := EIP.authenticationMethods()
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

	for _, c := range cases {
		m := methods.AuthenticationMethod[c.authmethod]
		r, err := m.preferredInnerAuthType(c.preferred)
		es := utils.ErrorString(err)
		if r.String() != c.want || es != c.err {
			t.Fatalf("Result: %s, %s Want: %s, %s", r, es, c.want, c.err)
		}
	}
}
