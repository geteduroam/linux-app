// package eap implements an XML eap-config parser compliant with the XML schema found at https://github.com/GEANT/CAT/blob/master/devices/eap_config/eap-metadata.xsd
// A part of this was generated with xgen https://github.com/xuri/xgen
// By hand modified:
// - Use NonEAPAuthNumbers as alias instead of hardcoded int
// - Removed -Properties
package eap

import (
	"encoding/xml"
	"errors"
	"fmt"

	"github.com/jwijenbergh/geteduroam-linux/internal/network"
	"github.com/jwijenbergh/geteduroam-linux/internal/network/inner"
	"github.com/jwijenbergh/geteduroam-linux/internal/network/method"
)

// VendorSpecificExtension ...
type VendorSpecificExtension struct {
	VendorAttr int `xml:"vendor,attr"`
}

// TypeSpecificExtension ...
type TypeSpecificExtension struct{}

// EAPMethod ...
type EAPMethod struct {
	Type           int                        `xml:"Type"`
	TypeSpecific   *TypeSpecificExtension     `xml:"TypeSpecific"`
	VendorSpecific []*VendorSpecificExtension `xml:"VendorSpecific"`
}

// NonEAPAuthNumbers is MSCHAPv2
type NonEAPAuthNumbers = int

// IEEE80211RSNProtocols is CTR with CBC-MAC Protocol (if used, only crypto setting
//
//	"WPA2/AES" and possible future protos are acceptable).
type IEEE80211RSNProtocols string

// NonEAPAuthMethod ...
type NonEAPAuthMethod struct {
	Type           NonEAPAuthNumbers          `xml:"Type"`
	TypeSpecific   *TypeSpecificExtension     `xml:"TypeSpecific"`
	VendorSpecific []*VendorSpecificExtension `xml:"VendorSpecific"`
}

// CertData ...
type CertData struct {
	FormatAttr   string `xml:"format,attr"`
	EncodingAttr string `xml:"encoding,attr"`
	Value        string `xml:",chardata"`
}

// LogoData ...
type LogoData struct {
	MimeAttr     string `xml:"mime,attr"`
	EncodingAttr string `xml:"encoding,attr"`
	Value        string `xml:",chardata"`
}

// ClientCredentialVariants is Not all EAP types and non-EAP authentication methods need or
//
//	support all types of credentials in the list below. While the
//	Schema allows to put all kinds of credential information inside
//	every AuthenticationMethod, even where the information is not
//	applicable, tags which are not applicable for an authentication
//	EAP or non-EAP type
//	   SHOULD NOT be included in the corresponding instance of
//	     AuthenticationMethod or InnerAuthenticationMethod when
//	     producing the XML file, and
//	   MUST be ignored by the entity consuming the XML file if
//	     present in the XML file.
type ClientCredentialVariants struct {
	AllowsaveAttr       bool   `xml:"allow_save,attr,omitempty"`
	OuterIdentity       string `xml:"OuterIdentity"`
	InnerIdentityPrefix string `xml:"InnerIdentityPrefix"`
	InnerIdentitySuffix string `xml:"InnerIdentitySuffix"`
	InnerIdentityHint   bool   `xml:"InnerIdentityHint"`
	// TODO: support `Username` (notice the n lowercase)?
	// See: https://github.com/geteduroam/windows-app/pull/39
	// Probably not needed as that pr states it is still not used
	// But maybe good to define here so we can check what is all in the xml?
	UserName                  string      `xml:"UserName"`
	Password                  string      `xml:"Password"`
	ClientCertificate         *CertData   `xml:"ClientCertificate"`
	IntermediateCACertificate []*CertData `xml:"IntermediateCACertificate"`
	Passphrase                string      `xml:"Passphrase"`
	PAC                       string      `xml:"PAC"`
	ProvisionPAC              bool        `xml:"ProvisionPAC"`
}

// ServerCredentialVariants is Not all EAP types and non-EAP authentication methods need or
//
//	support all types of credentials in the list below. While the
//	Schema allows to put all kinds of credential information inside
//	every AuthenticationMethod, even where the information is not
//	applicable, tags which are not applicable for an authentication
//	EAP or non-EAP type
//	   SHOULD NOT be included in the corresponding instance of
//	     AuthenticationMethod or InnerAuthenticationMethod when
//	     producing the XML file, and
//	   MUST be ignored by the entity consuming the XML file if
//	     present in the XML file.
type ServerCredentialVariants struct {
	CA       []*CertData `xml:"CA"`
	ServerID []string    `xml:"ServerID"`
}

// HelpdeskDetailElements ...
type HelpdeskDetailElements struct {
	EmailAddress []*LocalizedInteractive    `xml:"EmailAddress"`
	WebAddress   []*LocalizedNonInteractive `xml:"WebAddress"`
	Phone        []*LocalizedInteractive    `xml:"Phone"`
}

// LocationElements ...
type LocationElements struct {
	Longitude string `xml:"Longitude"`
	Latitude  string `xml:"Latitude"`
}

// ProviderInfoElements ...
type ProviderInfoElements struct {
	DisplayName      []*LocalizedNonInteractive `xml:"DisplayName"`
	Description      []*LocalizedNonInteractive `xml:"Description"`
	ProviderLocation []*LocationElements        `xml:"ProviderLocation"`
	ProviderLogo     *LogoData                  `xml:"ProviderLogo"`
	TermsOfUse       []*LocalizedNonInteractive `xml:"TermsOfUse"`
	Helpdesk         *HelpdeskDetailElements    `xml:"Helpdesk"`
}

// IEEE80211 is The conditions inside this element are considered AND conditions.
//
//	It does e.g. not make sense to have multiple SSIDs in one
//	IEEE80211 field because the condition would never
//	match. To specify multiple ORed network properties, use multiple
//	IEEE80211 instances.
type IEEE80211 struct {
	XMLName       xml.Name `xml:"IEEE80211"`
	SSID          string   `xml:"SSID"`
	ConsortiumOID string   `xml:"ConsortiumOID"`
	MinRSNProto   string   `xml:"MinRSNProto"`
}

// IEEE8023 ...
type IEEE8023 struct {
	XMLName   xml.Name `xml:"IEEE8023"`
	NetworkID string   `xml:"NetworkID"`
}

// CredentialApplicabilityType ...
type CredentialApplicabilityType struct {
	IEEE80211 []*IEEE80211 `xml:"IEEE80211"`
	IEEE8023  []*IEEE8023  `xml:"IEEE8023"`
}

// LocalizedInteractive ...
type LocalizedInteractive struct {
	LangAttr string `xml:"lang,attr,omitempty"`
	Value    string `xml:",chardata"`
}

// LocalizedNonInteractive ...
type LocalizedNonInteractive struct {
	LangAttr string `xml:"lang,attr,omitempty"`
	Value    string `xml:",chardata"`
}

// InnerAuthenticationMethod ...
type InnerAuthenticationMethod struct {
	EAPMethod            *EAPMethod                `xml:"EAPMethod"`
	NonEAPAuthMethod     *NonEAPAuthMethod         `xml:"NonEAPAuthMethod"`
	ServerSideCredential *ServerCredentialVariants `xml:"ServerSideCredential"`
	ClientSideCredential *ClientCredentialVariants `xml:"ClientSideCredential"`
}

// AuthenticationMethod ...
type AuthenticationMethod struct {
	EAPMethod                 *EAPMethod                   `xml:"EAPMethod"`
	ServerSideCredential      *ServerCredentialVariants    `xml:"ServerSideCredential"`
	ClientSideCredential      *ClientCredentialVariants    `xml:"ClientSideCredential"`
	InnerAuthenticationMethod []*InnerAuthenticationMethod `xml:"InnerAuthenticationMethod"`
}

// AuthenticationMethods ...
type AuthenticationMethods struct {
	AuthenticationMethod []*AuthenticationMethod `xml:"AuthenticationMethod"`
}

// EAPIdentityProvider ...
type EAPIdentityProvider struct {
	IDAttr                  string                       `xml:"ID,attr"`
	NamespaceAttr           string                       `xml:"namespace,attr"`
	VersionAttr             int                          `xml:"version,attr,omitempty"`
	LangAttr                string                       `xml:"lang,attr,omitempty"`
	ValidUntil              string                       `xml:"ValidUntil"`
	AuthenticationMethods   *AuthenticationMethods       `xml:"AuthenticationMethods"`
	CredentialApplicability *CredentialApplicabilityType `xml:"CredentialApplicability"`
	ProviderInfo            *ProviderInfoElements        `xml:"ProviderInfo"`
	VendorSpecific          *VendorSpecificExtension     `xml:"VendorSpecific"`
}

// EAPIdentityProviderList ...
type EAPIdentityProviderList struct {
	EAPIdentityProvider *EAPIdentityProvider `xml:"EAPIdentityProvider"`
}

// Parse parses a byte array into the main EAPIdentityProviderList struct. It returns nil if error
func Parse(data []byte) (*EAPIdentityProviderList, error) {
	var eap EAPIdentityProviderList
	err := xml.Unmarshal(data, &eap)
	if err != nil {
		return nil, err
	}
	return &eap, nil
}

func (p *EAPIdentityProvider) authenticationMethods() (*AuthenticationMethods, error) {
	m := p.AuthenticationMethods
	if m == nil {
		return nil, errors.New("authentication methods section couldn't be found")
	}
	return m, nil
}

func (p *EAPIdentityProvider) AuthMethods() ([]*AuthenticationMethod, error) {
	ams, err := p.authenticationMethods()
	if err != nil {
		return nil, fmt.Errorf("failed to get authentication method due to no methods available: %v", err)
	}
	am := ams.AuthenticationMethod
	if len(am) < 1 {
		return nil, errors.New("authentication method couldn't be found")
	}
	return am, nil
}

// preferredInnerAuthType gets the first valid inner authentication type
func (am *AuthenticationMethod) preferredInnerAuthType(mt method.Type) (inner.Type, error) {
	if len(am.InnerAuthenticationMethod) < 1 {
		return inner.NONE, errors.New("inner authentication method couldn't be found")
	}

	// loop through all methods and return the first valid one
	for _, i := range am.InnerAuthenticationMethod {
		if i.EAPMethod != nil {
			if inner.Valid(mt, i.EAPMethod.Type, true) {
				return inner.Type(i.EAPMethod.Type), nil
			}
		}

		// Otherwise try to get Non eap
		if i.NonEAPAuthMethod != nil {
			if inner.Valid(mt, i.NonEAPAuthMethod.Type, false) {
				return inner.Type(i.NonEAPAuthMethod.Type), nil
			}
		}
	}
	return inner.NONE, errors.New("no viable inner authentication method found")
}

// SSID gets the SSID from the eap identity provider
// SSIDSettings returns the SSID and MinRSNProto associated with it
// It loops trough the credential applicability list and gets the first candidate
// The candidate filtering was based on https://github.com/geteduroam/windows-app/blob/f11f00dee3eb71abd38537e18881463f83b180d3/EduroamConfigure/EapConfig.cs#L84
func (p *EAPIdentityProvider) SSIDSettings() (string, string, error) {
	if p.CredentialApplicability == nil {
		return "", "", errors.New("no Credential Applicability found")
	}
	if len(p.CredentialApplicability.IEEE80211) < 1 {
		return "", "", errors.New("no IEE80211 section found")
	}
	for _, i := range p.CredentialApplicability.IEEE80211 {
		if i == nil {
			continue
		}

		// no min rsn proto
		if i.MinRSNProto == "" {
			continue
		}

		// no ssid present
		if i.SSID == "" {
			continue
		}

		// tkip is too insecure
		if i.MinRSNProto == "TKIP" {
			continue
		}

		return i.SSID, i.MinRSNProto, nil
	}
	return "", "", errors.New("no viable SSID entry found")
}

func (ca *CertData) Valid() bool {
	if ca.EncodingAttr != "base64" {
		return false
	}

	if ca.FormatAttr != "X.509" {
		return false
	}

	return true
}

func (ss *ServerCredentialVariants) CAList() ([]string, error) {
	var ca []string
	for _, c := range ss.CA {
		if c != nil && c.Valid() {
			ca = append(ca, c.Value)
		}
	}
	if len(ca) == 0 {
		return nil, errors.New("no viable server side CA entry found")
	}
	return ca, nil
}

// TLSNetwork creates a TLS network using the authentication method
func (m *AuthenticationMethod) TLSNetwork(base network.Base) network.Network {
	// TODO: client certificate should be required right?
	var ccert string
	var passphrase string
	if cc := m.ClientSideCredential; cc != nil {
		if cc.ClientCertificate != nil && cc.ClientCertificate.Valid() {
			ccert = cc.ClientCertificate.Value
		}
		passphrase = cc.Passphrase
	}

	return &network.TLS{
		Base:              base,
		ClientCertificate: ccert,
		Password:          passphrase,
	}
}

// NonTLSNetwork creates a network that is Non-TLS using the authentication method
func (m *AuthenticationMethod) NonTLSNetwork(base network.Base) (network.Network, error) {
	// Define defaults
	var username, password, identity, prefix, suffix string
	if cc := m.ClientSideCredential; cc != nil {
		username = cc.UserName
		password = cc.Password
		identity = cc.OuterIdentity
		// TODO: Can this be false but still hints here?
		if cc.InnerIdentityHint {
			prefix = cc.InnerIdentityPrefix
			if cc.InnerIdentitySuffix != "" {
				suffix = fmt.Sprintf("@%s", cc.InnerIdentitySuffix)
			}
		}

	}
	mt := method.Type(m.EAPMethod.Type)
	// get the inner auth
	it, err := m.preferredInnerAuthType(mt)
	if err != nil {
		return nil, errors.New("no preferred inner authentication found")
	}

	return &network.NonTLS{
		Base:         base,
		Username:     username,
		Prefix:       prefix,
		Suffix:       suffix,
		Password:     password,
		MethodType:   method.Type(m.EAPMethod.Type),
		InnerAuth:    it,
		AnonIdentity: identity,
	}, nil
}

// Network gets a network for an authentication method
func (m *AuthenticationMethod) Network(ssid string, minrsn string) (network.Network, error) {
	// We check if the eap method is valid
	if m.EAPMethod == nil || !method.Valid(m.EAPMethod.Type) {
		return nil, errors.New("no EAP method")
	}
	mt := m.EAPMethod.Type

	// Get the server side credentials
	ss := m.ServerSideCredential
	if ss == nil {
		return nil, errors.New("no server side credentials")
	}

	CA, err := ss.CAList()
	if err != nil {
		return nil, errors.New("no preferred server side CA found")
	}

	// Create the Base
	// These are the settings that are common for each network
	sid := ss.ServerID
	base := network.Base{
		Cert:      CA,
		SSID:      ssid,
		MinRSN:    minrsn,
		ServerIDs: sid,
	}

	// If TLS we need to construct different arguments than when we have Non TLS
	if method.Type(mt) == method.TLS {
		return m.TLSNetwork(base), nil
	} else {
		return m.NonTLSNetwork(base)
	}
}

// Network creates a TLS or NON-TLS secured network from the EAP config
// This network can then afterwards be imported into NetworkManager
func (eap *EAPIdentityProviderList) Network() (network.Network, error) {
	// Get the provider section
	p := eap.EAPIdentityProvider
	if p == nil {
		return nil, errors.New("identity provider section couldn't be found")
	}
	methods, err := p.AuthMethods()
	if err != nil {
		return nil, err
	}
	ssid, minrsn, err := p.SSIDSettings()
	if err != nil {
		return nil, err
	}
	// TODO: clean this big boy up
	for _, m := range methods {
		n, err := m.Network(ssid, minrsn)
		if err != nil {
			// TODO: log error
			continue
		}
		return n, nil
	}
	return nil, errors.New("no viable network settings found in EAP config")
}
