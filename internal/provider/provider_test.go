package provider

import (
	"testing"

	"github.com/geteduroam/linux-app/internal/utilsx"
	"golang.org/x/text/language"
)

func TestLocalizedStrings(t *testing.T) {
	cases := []struct {
		input LocalizedStrings
		lang  language.Tag
		want  string
	}{
		{
			input: LocalizedStrings{
				{Display: "disp_en", Lang: "en"},
				{Display: "disp_nl", Lang: "nl"},
			},
			lang: language.English,
			want: "disp_en",
		},
		{
			input: LocalizedStrings{
				{Display: "disp_en", Lang: "en"},
				{Display: "disp_nl", Lang: "nl"},
			},
			lang: language.Dutch,
			want: "disp_nl",
		},
		{
			input: LocalizedStrings{
				{Display: "disp_en", Lang: "en"},
			},
			lang: language.German,
			want: "disp_en",
		},
		{
			input: LocalizedStrings{
				{Display: "disp_en", Lang: "en"},
				{Display: "disp_gb", Lang: "en_GB"},
			},
			lang: language.BritishEnglish,
			want: "disp_gb",
		},
		{
			input: LocalizedStrings{
				{Display: "disp_en", Lang: "en"},
				{Display: "disp_gb", Lang: "en_GB"},
			},
			lang: language.English,
			want: "disp_en",
		},
	}

	for _, c := range cases {
		systemLanguage = c.lang
		got := c.input.Get()
		if got != c.want {
			t.Fatalf("Got: %s, Not equal to Want: %s", got, c.want)
		}
	}
}

func TestFilterSort(t *testing.T) {
	i := Providers{
		{
			Name: LocalizedStrings{{Display: "Provider One"}},
		},
		{
			// Diacritics
			Name: LocalizedStrings{{Display: "Provider Twö"}},
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
			want:   "Provider One",
		},
		{
			// Filter case-insensitive
			input:  "one",
			length: 1,
			want:   "Provider One",
		},
		{
			// Filter case-insensitive diacriticless
			input:  "two",
			length: 1,
			want:   "Provider Twö",
		},
		{
			// Filter all case-insensitive diacriticless
			input:  "provider",
			length: 2,
			want:   "Provider Twö",
		},
	}

	for _, c := range cases {
		result := i.FilterSort(c.input)
		length := len(*result)
		name := (*result)[0].Name
		if name.Get() != c.want || length != c.length {
			t.Fatalf("Result: %s, %d, Want: %s, %d", name, length, c.want, c.length)
		}
	}
}

func TestFlow(t *testing.T) {
	p := Profile{
		EapConfigEndpoint:    "https://provider1.geteduroam.nl/api/eap-config/",
		MobileConfigEndpoint: "https://provider1.geteduroam.nl/api/eap-config/?format=mobileconfig",
		WebviewEndpoint:      "https://provider1.geteduroam.nl/",
		ID:                   "letswifi_cat_0001",
		Name:                 LocalizedStrings{{Display: "geteduroam"}},
		Type:                 "webview",
	}

	var flow FlowCode
	flow = p.Flow()
	if flow != RedirectFlow {
		t.Fatalf("Flow should be RedirectFlow")
	}

	p.Type = ""
	flow = p.Flow()
	if flow != OAuthFlow {
		t.Fatalf("Flow should be OAuthFlow")
	}

	p.Type = "eap-config"
	flow = p.Flow()
	if flow != DirectFlow {
		t.Fatalf("Flow should be DirectFlow")
	}
}

func TestRedirectURI(t *testing.T) {
	p := Profile{
		ID:              "letswifi_cat_0001",
		Name:            LocalizedStrings{{Display: "geteduroam"}},
		WebviewEndpoint: "",
	}

	cases := []struct {
		input string
		want  string
		e     string
	}{
		{
			// Normal test
			input: "https://provider1.geteduroam.nl/",
			want:  "https://provider1.geteduroam.nl/",
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
			input: "http://provider1.geteduroam.nl/",
			want:  "https://provider1.geteduroam.nl/",
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
		p.WebviewEndpoint = c.input
		r, e := p.RedirectURI()
		es := utilsx.ErrorString(e)
		if r != c.want || es != c.e {
			t.Fatalf("Result: %s, %s Want: %s, %s", r, es, c.want, c.e)
		}
	}
}
