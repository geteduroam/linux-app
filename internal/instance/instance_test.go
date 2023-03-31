package instance

import (
	"testing"
)

func Test_Filter(t *testing.T) {
	i := Instances{}

	i = []Instance {
		Instance {
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
		Instance {
			// Diacritics
			Name:         "Instånce Twö",
		},
	}

	cases := []struct {
		input string
		want  string
	}{
		{
			// Normal test
			input: "One",
			want:  "Instance One",
		},
		{
			// Filter case-insensitive
			input: "one",
			want:  "Instance One",
		},
		{
			// Filter case-insensitive diacriticless
			input: "two",
			want:  "Instånce Twö",
		},
	}

	for _, c := range cases {
		result := i.Filter(c.input)
		name := (*result)[0].Name
		if name != c.want {
			t.Fatalf("Result: %s, Want: %s", name, c.want)
		}
	}
}
