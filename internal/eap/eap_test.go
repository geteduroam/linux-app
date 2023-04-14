package eap

import (
	"os"
	"testing"

	"github.com/geteduroam/linux-app/internal/network/inner"
	"github.com/geteduroam/linux-app/internal/network/method"
	"github.com/geteduroam/linux-app/internal/utils"
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
		method method.Type
		want   inner.Type
		err    string
	}{
		// The first autentication method only has 26 defined EapMschapv2
		// This is a valid method for PEAP
		{
			method: method.PEAP,
			want:   inner.EapMschapv2,
			err:    "",
		},
		// We want 26, EapMschapv2 again due to it only having 26 defined again
		{
			method: method.PEAP,
			want:   inner.EapMschapv2,
			err:    "",
		},
		// This only has a non eap auth method so no viable is found here
		{
			method: method.PEAP,
			want:   inner.None,
			err:    "no viable inner authentication method found",
		},
	}

	for i, c := range cases {
		m := methods.AuthenticationMethod[i]
		r, err := m.preferredInnerAuthType(c.method)
		if r != c.want {
			t.Fatalf("method is not what is expected, got: %d, want: %d", r, c.want)
		}
		if utils.ErrorString(err) != c.err {
			t.Fatalf("error is not expected, got: %v, want: %v", err, c.err)
		}
	}
}
