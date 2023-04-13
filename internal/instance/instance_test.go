package instance

import (
	"testing"
)

func Test_Filter(t *testing.T) {
	i := []Instance {
		{
			Name:         "Instance One",
			/*
			For the moment we can do without these since
			filtering only takes place on Name, but we might
			need them later if Filtering logic changes

			CatIDP:       0001,
			Country:      "NL",
			Geo: []geo {
				{
					lat:  float32(0),
					long: float32(0),
				},

			},
			ID:           "cat_0001",
			Profiles: []Profile {
				{
					AuthorizationEndpoint: "https://instance1.geteduroam.nl/oauth/authorize/",
					Default:               true,
					EapConfigEndpoint:     "https://instance1.geteduroam.nl/api/eap-config/",
					ID:                    "letswifi_cat_0001",
					Name:                  "geteduroam",
					OAuth:                 true,
					TokenEndpoint:         "https://instance1.geteduroam.nl/oauth/token/",
					Redirect:              "https://instance1.geteduroam.nl/",
				},
			},
			*/
		},
		{
			// Diacritics
			Name:         "Instånce Twö",
		},
	}

	cases := []struct {
		input string
		length int
		want  string
	}{
		{
			// Normal test
			input: "One",
			length: 1,
			want:  "Instance One",
		},
		{
			// Filter case-insensitive
			input: "one",
			length: 1,
			want:  "Instance One",
		},
		{
			// Filter case-insensitive diacriticless
			input: "two",
			length: 1,
			want:  "Instånce Twö",
		},
		{
			// Filter all case-insensitive diacriticless
			input: "instance",
			length: 2,
			want:  "Instance One",
		},
	}

	for _, c := range cases {
		result := i.Filter(c.input)
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
	EtoString := func(e error) string {
		if e != nil {
			return e.Error()
		}
		return ""
	}

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
			e:     "No redirect found",
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
		es := EtoString(e)
		if r != c.want || es != c.e {
			t.Fatalf("Result: %s, %s Want: %s, %s", r, es, c.want, c.e)
		}
	}
}
