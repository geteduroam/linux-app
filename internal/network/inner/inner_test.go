package inner

import (
	"testing"

	"github.com/geteduroam/linux-app/internal/network/method"
)

func Test_IsValid(t *testing.T) {
	cases := []struct {
		mt    method.Type
		input int
		eap   bool
		want  bool
	}{
		// We have an eap type but the inner is not an eap type
		{
			input: 3, // 3: PAP
			eap:   true,
			want:  false,
		},
		// We have as method TLS, we should accept anything as it's not used anyways
		{
			mt:    method.TLS,
			input: 0,
			want:  true,
		},
		{
			mt:    method.TLS,
			input: 3,
			want:  true,
		},
		{
			mt:    method.TLS,
			input: 25,
			want:  true,
		},
		{
			mt:    method.TLS,
			input: 50,
			want:  true,
		},
		// TTLS we support different types than with PEAP
		{
			mt:    method.TTLS,
			input: 3, // 3: PAP
			eap:   false,
			want:  true,
		},
		{
			mt:    method.TTLS,
			input: 25, // 25: EAP_PEAP_MSCHAPV2
			eap:   true,
			want:  false,
		},
		{
			mt:    method.TTLS,
			input: 26, // 26: EAP_PEAP_MSCHAPV2
			eap:   true,
			want:  true,
		},
		{
			mt:    method.TTLS,
			input: 27, // 27: bogus
			eap:   false,
			want:  false,
		},
		{
			mt:    method.PEAP,
			input: 25, // 25: EAP_PEAP_MSCHAPV2
			eap:   true,
			want:  true,
		},
		{
			mt:    method.PEAP,
			input: 25,    // 25: EAP_PEAP_MSCHAPV2
			eap:   false, // EAP not mathcing
			want:  false,
		},
		{
			mt:    method.PEAP,
			input: 26, // 25: EAP_MSCHAPV2
			eap:   true,
			want:  true,
		},
		{
			mt:    method.PEAP,
			input: 26,    // 26: EAP_MSCHAPV2
			eap:   false, // EAP not mathcing
			want:  false,
		},
	}

	for _, c := range cases {
		got := IsValid(c.mt, c.input, c.eap)
		if got != c.want {
			t.Fatalf("Got: %v, Want: %v, when testing method type: %v, input: %v, eap: %v", got, c.want, c.mt, c.input, c.eap)
		}
	}
}
