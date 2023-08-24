package instance

import (
	"testing"

	"github.com/geteduroam/linux-app/internal/utils"
)

func Test_FilterSort(t *testing.T) {
	i := Instances{
		{
			Name: "Instance One",
		},
		{
			// Diacritics
			Name: "Instånce Twö",
		},
	}

	cases := []struct {
		input  string
		length int
		want   string
	}{
		{
			// Normal test
			input:  "One",
			length: 1,
			want:   "Instance One",
		},
		{
			// Filter case-insensitive
			input:  "one",
			length: 1,
			want:   "Instance One",
		},
		{
			// Filter case-insensitive diacriticless
			input:  "two",
			length: 1,
			want:   "Instånce Twö",
		},
		{
			// Filter all case-insensitive diacriticless
			input:  "instance",
			length: 2,
			want:   "Instånce Twö",
		},
	}

	for _, c := range cases {
		result := i.FilterSort(c.input)
		length := len(*result)
		name := (*result)[0].Name
		if name != c.want || length != c.length {
			t.Fatalf("Result: %s, %d, Want: %s, %d", name, length, c.want, c.length)
		}
	}
}

func Test_Flow(t *testing.T) {
	p := Profile{
		AuthorizationEndpoint: "https://instance1.geteduroam.nl/oauth/authorize/",
		Default:               true,
		EapConfigEndpoint:     "https://instance1.geteduroam.nl/api/eap-config/",
		ID:                    "letswifi_cat_0001",
		Name:                  "geteduroam",
		OAuth:                 true,
		TokenEndpoint:         "https://instance1.geteduroam.nl/oauth/token/",
		Redirect:              "https://instance1.geteduroam.nl/",
	}

	var flow FlowCode
	flow = p.Flow()
	if flow != RedirectFlow {
		t.Fatalf("Flow should be RedirectFlow")
	}

	p.Redirect = ""
	flow = p.Flow()
	if flow != OAuthFlow {
		t.Fatalf("Flow should be OAuthFlow")
	}

	p.OAuth = false
	flow = p.Flow()
	if flow != DirectFlow {
		t.Fatalf("Flow should be DirectFlow")
	}
}

func Test_RedirectURI(t *testing.T) {
	p := Profile{
		AuthorizationEndpoint: "https://instance1.geteduroam.nl/oauth/authorize/",
		Default:               true,
		EapConfigEndpoint:     "https://instance1.geteduroam.nl/api/eap-config/",
		ID:                    "letswifi_cat_0001",
		Name:                  "geteduroam",
		OAuth:                 true,
		TokenEndpoint:         "https://instance1.geteduroam.nl/oauth/token/",
		Redirect:              "",
	}

	cases := []struct {
		input string
		want  string
		e     string
	}{
		{
			// Normal test
			input: "https://instance1.geteduroam.nl/",
			want:  "https://instance1.geteduroam.nl/",
			e:     "",
		},
		{
			// No Redirect
			input: "",
			want:  "",
			e:     "no redirect found",
		},
		{
			// Enforce Test
			input: "http://instance1.geteduroam.nl/",
			want:  "https://instance1.geteduroam.nl/",
			e:     "",
		},
		{
			// No URL
			input: "foobar",
			want:  "https://foobar",
			e:     "",
		},
	}

	for _, c := range cases {
		p.Redirect = c.input
		r, e := p.RedirectURI()
		es := utils.ErrorString(e)
		if r != c.want || es != c.e {
			t.Fatalf("Result: %s, %s Want: %s, %s", r, es, c.want, c.e)
		}
	}
}
