package eap

import (
	"os"
	"path"
	"reflect"
	"testing"

	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/network/cert"
	"github.com/geteduroam/linux-app/internal/network/inner"
	"github.com/geteduroam/linux-app/internal/network/method"
	"github.com/geteduroam/linux-app/internal/utils"
)

type authMethodTest struct {
	want inner.Type
	err  string
}

func testAuthMethod(t *testing.T, eip *EAPIdentityProvider, cases []authMethodTest) {
	methods, err := eip.authenticationMethods()
	if err != nil {
		t.Fatalf("failed getting authentication methods: %v", err)
	}

	for i, c := range cases {
		m := methods.AuthenticationMethod[i]
		r, err := m.preferredInnerAuthType()
		if r != c.want {
			t.Fatalf("method is not what is expected, got: %d, want: %d", r, c.want)
		}
		if utils.ErrorString(err) != c.err {
			t.Fatalf("error is not expected, got: %v, want: %v", err, c.err)
		}
	}
}

func testProviderInfo(t *testing.T, eip *EAPIdentityProvider, pi network.ProviderInfo) {
	got := eip.PInfo()
	if !reflect.DeepEqual(pi, got) {
		t.Fatalf("provider info not equal, want: %v, got: %v", pi, got)
	}
}

type ssidSettingsTest struct {
	SSIDs []network.SSID
	err   string
}

func testSSIDSettings(t *testing.T, eip *EAPIdentityProvider, settings ssidSettingsTest) {
	gotSSIDs, gotErr := eip.SSIDSettings()
	if !reflect.DeepEqual(gotSSIDs, settings.SSIDs) {
		t.Fatalf("SSIDs is not equal, got: %v, want: %v", gotSSIDs, settings.SSIDs)
	}
	gotErrS := utils.ErrorString(gotErr)
	if gotErrS != settings.err {
		t.Fatalf("Error for SSID settings is not equal, got: %v, want: %v", gotErrS, settings.err)
	}
}

type networkTest struct {
	n   network.Network
	err string
}

type parseTest struct {
	filename         string
	authMethodTests  []authMethodTest
	providerInfoTest network.ProviderInfo
	ssidTest         ssidSettingsTest
	netTest          networkTest
}

func mustParseCert(t *testing.T, root string) cert.Certs {
	c, err := cert.New([]string{root})
	if err != nil {
		t.Fatalf("failed to generate certs: %v", err)
	}
	return *c
}

func TestParse(t *testing.T) {
	tests := []parseTest{
		{
			filename: "eva-eap.xml",
			authMethodTests: []authMethodTest{
				// In this file we expect everything to be valid so errors are nil
				// The first authentication method, PEAP, only has EapMschapv2 (26) as inner defined
				{
					want: inner.EapMschapv2,
					err:  "",
				},
				// The second authentication method, 21 (TTLS), only has 26 again, EapMschapv2
				{
					want: inner.EapMschapv2,
					err:  "",
				},
				// The third authentication method, 21 TTLS, only has a Non EAP Auth method 1
				{
					want: inner.Pap,
					err:  "",
				},
			},
			providerInfoTest: network.ProviderInfo{
				Name:        "eduroam Visitor Access (eVA)",
				Description: "eVA",
			},
			ssidTest: ssidSettingsTest{
				SSIDs: []network.SSID{{
					Value:  "eduroam",
					MinRSN: "CCMP",
				}},
				err: "",
			},
			netTest: networkTest{
				n: &network.NonTLS{
					Base: network.Base{
						AnonIdentity: "anonymous@edu.nl",
						Certs: mustParseCert(t,
							"MIIDtzCCAp+gAwIBAgIUCVQbKTO9PsqghECzGPqq6Fiy8REwDQYJKoZIhvcNAQELBQAwazELMAkGA1UEBhMCTkwxEzARBgNVBAgMClNvbWUtU3RhdGUxEjAQBgNVBAcMCUFtc3RlcmRhbTEQMA4GA1UECgwHVGVzdGluZzENMAsGA1UECwwEVGVzdDESMBAGA1UEAwwJVGVzdCB0ZXN0MB4XDTIzMDUyNDEzNTUxMFoXDTMzMDUyMTEzNTUxMFowazELMAkGA1UEBhMCTkwxEzARBgNVBAgMClNvbWUtU3RhdGUxEjAQBgNVBAcMCUFtc3RlcmRhbTEQMA4GA1UECgwHVGVzdGluZzENMAsGA1UECwwEVGVzdDESMBAGA1UEAwwJVGVzdCB0ZXN0MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyLqG9yuMhbVC5y9zofPDLeCDIUVjgPbxXHtM6uveBUtqG4PxDkTczOlYN1IsYRh2iLNRYY4cqYZ1qtW+1CZaFVowhUMbTR7Y8Ik10CrCJQqoGq1CIICBd50wTFBLU2MZU3LQTwKYb5VQgbCMvRVHWdQOYg5GSlgdJRtIbzV1d+Q7+N5jiEBsT6psSu2gBduF1ueGICKe6Fk+ckOHDpwjVGeNIxnN2hJ5ft3WReDJ7fcHLMx7lNS+ZeY35LtpYiT6I8RGlMh2bu9hMTY1jXNbEqqZ2/5TmjVygS7BEMrVage9K2I5eM8++yX27OV3Di/SM3q/RVIcu1lNKaSj0IxXhwIDAQABo1MwUTAdBgNVHQ4EFgQU0M2QAnLWEDSFdFLCm5OxvVA9D1swHwYDVR0jBBgwFoAU0M2QAnLWEDSFdFLCm5OxvVA9D1swDwYDVR0TAQH/BAUwAwEB/zANBgkqhkiG9w0BAQsFAAOCAQEAHHdxGNUmyZa4ER9oqSalwVy9W5y1cNr4VpxBbxJe/fBPp+xdtnYRbz1/93LwcA+bTJlvT8ez2ijOJj5QODrgeVy5r4p5/1cABnJhsszk6ffJy/n5vIqo9jp8+7ZTFGxm1QQAOoZfJM+3ft8ZFf5e8Vjh090QV2OZvV69sey+TvfAlNMVotf/CaA2zA/j4z2bmWdrLAc5VVrb1Mil4z7LHhL62oOwXrS85zuoVBQVMbh5tnYgzMnbuy0hmMDg3ClkmSQTqzPyEi0SjhqKjgLgyVa47myhxvr1y77k0rZBRzkSEMsopu+ANYoVKRpw7gmjgMmXWzvdNlbD6RgpGlR4iA=="),
						SSIDs: []network.SSID{{
							Value:  "eduroam",
							MinRSN: "CCMP",
						}},
						ServerIDs: []string{
							"edu.nl",
						},
						ProviderInfo: network.ProviderInfo{
							Name:        "eduroam Visitor Access (eVA)",
							Description: "eVA",
						},
					},
					Credentials: network.Credentials{
						Suffix: "@edu.nl",
					},
					MethodType: method.PEAP,
					InnerAuth:  inner.EapMschapv2,
				},
				err: "",
			},
		},
		{
			// changes:
			// - removed provider info
			// - changed auth methods to contain some invalid values
			// - ssid entry removed
			filename: "eva-eap-changed.xml",
			authMethodTests: []authMethodTest{
				// The first authentication method, PEAP, has no inners defined
				{
					want: inner.None,
					err:  "the authentication method has no inner authentication methods",
				},
				// The second authentication method, also PEAP, has changed inner type to Non EAP
				{
					want: inner.None,
					err:  "no viable inner authentication method found",
				},
				// The third authentication method, PEAP, has changed inner type to Non EAP
				{
					want: inner.Mschap,
					err:  "",
				},
			},
			providerInfoTest: network.ProviderInfo{},
			ssidTest: ssidSettingsTest{
				err: "no viable SSID entries found",
			},
			netTest: networkTest{
				n:   nil,
				err: "no viable SSID entries found",
			},
		},
	}

	for _, c := range tests {
		b, err := os.ReadFile(path.Join("test_data", c.filename))
		if err != nil {
			t.Fatalf("failed reading file: %v", err)
		}

		eipl, err := Parse(b)
		if err != nil {
			t.Fatalf("failed parsing file: %v", err)
		}
		eip := eipl.EAPIdentityProvider
		if eip == nil {
			t.Fatalf("no eap identity provider found")
		}

		// test the individual components that make up the network
		testAuthMethod(t, eip, c.authMethodTests)
		testProviderInfo(t, eip, c.providerInfoTest)
		testSSIDSettings(t, eip, c.ssidTest)

		// finally test the whole network we get back
		n, err := eipl.Network()
		errS := utils.ErrorString(err)
		if errS != c.netTest.err {
			t.Fatalf("network error not equal. Got: %v, want: %v", errS, c.netTest.err)
		}
		if !reflect.DeepEqual(n, c.netTest.n) {
			t.Fatalf("networks are not equal. Got: %v, want: %v", n, c.netTest.n)
		}
	}
}

func TestCertFromContainer(t *testing.T) {
	cases := []struct {
		// valid data will be generated with test_data/genpkcs12.sh
		cert       string
		passphrase string
		hasErr     bool
		wantNil    bool
	}{
		{
			// a valid PKCS12 encrypted with "" should have no error
			cert:       "pkcs12empty",
			passphrase: "",
			hasErr:     false,
			wantNil:    false,
		},
		{
			// a valid PKCS12 encrypted with "" should have an error if we pass passphrase test
			cert:       "pkcs12empty",
			passphrase: "test",
			hasErr:     true,
			wantNil:    true,
		},
		{
			// a valid PKCS12 encrypted with "test" should have no error if we give passphrase ""
			// This is because in this case we want to try again. It should however give nil data
			cert:       "pkcs12test",
			passphrase: "",
			hasErr:     false,
			wantNil:    true,
		},
		{
			// a valid PKCS12 encrypted with "test" should have no error if we give passphrase "test"
			cert:       "pkcs12test",
			passphrase: "test",
			hasErr:     false,
			wantNil:    false,
		},
		{
			// a valid PKCS12 encrypted with "test" should have an error if we give passphrase "test2"
			cert:       "pkcs12test",
			passphrase: "test2",
			hasErr:     true,
			wantNil:    true,
		},
		// an invalid PKCS12 should always have an error
		{
			cert:       "pkcs12invalid",
			passphrase: "",
			hasErr:     true,
			wantNil:    true,
		},
		{
			cert:       "pkcs12invalid",
			passphrase: "test",
			hasErr:     true,
			wantNil:    true,
		},
	}

	for idx, c := range cases {
		cert, err := os.ReadFile(path.Join("test_data", c.cert))
		if err != nil {
			t.Fatalf("Failed reading cert file: %v, idx: %v", err, idx)
		}
		g, gerr := certFromContainer(string(cert), c.passphrase)
		if c.hasErr != (gerr != nil) {
			t.Fatalf("Has error: %v, got error: %v, idx: %v", c.hasErr, gerr, idx)
		}
		// test if nil is always returned if we have an error
		if c.wantNil != (g == nil) {
			t.Fatalf("Want nil: %v, got result: %v, idx: %v", c.wantNil, g, idx)
		}
	}
}
