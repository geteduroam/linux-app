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

	"golang.org/x/exp/slog"

	"github.com/geteduroam/linux-app/internal/network"
	"github.com/geteduroam/linux-app/internal/network/cert"
	"github.com/geteduroam/linux-app/internal/network/inner"
	"github.com/geteduroam/linux-app/internal/network/method"
)

// VendorSpecificExtension ...
type VendorSpecificExtension struct {
	VendorAttr int `xml:"vendor,attr"`
}

// TypeSpecificExtension ...
type TypeSpecificExtension struct{}

// EAPMethod ...
type EAPMethod struct { //revive:disable-line:exported
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
type EAPIdentityProvider struct { //revive:disable-line:exported
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
type EAPIdentityProviderList struct { //revive:disable-line:exported
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

// authenticationMethods gets the authentication methods from the EAP identity provider
// It has additional checking for NULL pointers
func (p *EAPIdentityProvider) authenticationMethods() (*AuthenticationMethods, error) {
	m := p.AuthenticationMethods
	if m == nil {
		return nil, errors.New("authentication methods section couldn't be found")
	}
	return m, nil
}

// AuthMethods gets a list of authentication methods by checking if it is NON-NULL and NON-EMPTY
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
func (am *AuthenticationMethod) preferredInnerAuthType() (inner.Type, error) {
	if len(am.InnerAuthenticationMethod) < 1 {
		return inner.None, errors.New("the authentication method has no inner authentication methods")
	}

	mt := method.Type(am.EAPMethod.Type)

	// loop through all methods and return the first valid one
	for _, i := range am.InnerAuthenticationMethod {
		if i.EAPMethod != nil {
			if inner.IsValid(mt, i.EAPMethod.Type, true) {
				return inner.Type(i.EAPMethod.Type), nil
			}
		}

		// Otherwise try to get Non eap
		if i.NonEAPAuthMethod != nil {
			if inner.IsValid(mt, i.NonEAPAuthMethod.Type, false) {
				return inner.Type(i.NonEAPAuthMethod.Type), nil
			}
		}
	}
	return inner.None, errors.New("no viable inner authentication method found")
}

// SSIDSettings returns the first valid SSID and the MinRSNProto associated with it
// It loops trough the credential applicability list and gets the first valid candidate
// The candidate filtering was based on https://github.com/geteduroam/windows-app/blob/f11f00dee3eb71abd38537e18881463f83b180d3/EduroamConfigure/EapConfig.cs#L84
// A candidate is valid if:
//   - MinRSNProto is not empty, TODO: shouldn't we just default to CCMP?
//   - The SSID is not empty
//   - The MinRSNProto is NOT TKIP as that is insecure
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

// isValid returns whether or not a certificate is valid by checking if the encoding is base64 and the format is `format`
func (ca *CertData) isValid(format string) bool {
	if ca == nil {
		return false
	}
	if ca.EncodingAttr != "base64" {
		return false
	}

	if ca.FormatAttr != format {
		return false
	}

	return true
}

// CaList gets a list of certificates by looping through the certificate list and returning all *valid* certificates
func (ss *ServerCredentialVariants) CAList() (*cert.Certs, error) {
	var certs []string
	for _, c := range ss.CA {
		if c.isValid("X.509") {
			certs = append(certs, c.Value)
		}
	}
	if len(certs) == 0 {
		return nil, errors.New("no viable server side CA entry found")
	}
	return cert.New(certs)
}

// TLSNetwork creates a TLS network using the authentication method.
// The base that is passed here are settings that are common between TLS and NON-TLS networks
func (am *AuthenticationMethod) TLSNetwork(base network.Base) (network.Network, error) {
	var ccert string
	var passphrase string
	var identity string
	if csc := am.ClientSideCredential; csc != nil {
		cc := csc.ClientCertificate
		if cc.isValid("PKCS12") {
			ccert = cc.Value
			passphrase = csc.Passphrase
		}
		identity = csc.OuterIdentity
	}

	// there could be multiple things going wrong:
	// - The pkcs12 is not given and no passphrase is given
	// - The pkcs12 is not given but a passphrase is given
	// - The pkcs12 is given but the passphrase is empty
	// - The pkcs12 is given but the passphrase is wrong

	// the conditions that we will handle as an explicit error is:
	// - wrong passphrase
	// - wrong format of the ccert
	// NewClientCert will then give us an error

	var fcc *cert.ClientCert
	var err error
	// If we should not be asking for a certificate we can construct it now and return an explicit error if something went wrong
	if ccert != "" && passphrase != "" {
		// create the final client certificate structure
		fcc, err = cert.NewClientCert(ccert, passphrase)
		if err != nil {
			return nil, err
		}
	}

	base.AnonIdentity = identity
	return &network.TLS{
		Base:       base,
		ClientCert: fcc,
		RawPKCS12: ccert,
		Password:   passphrase,
	}, nil
}

// NonTLSNetwork creates a network that is Non-TLS using the authentication method
// The base that is passed here are settings that are common between TLS and NON-TLS networks
func (am *AuthenticationMethod) NonTLSNetwork(base network.Base) (network.Network, error) {
	// Define defaults
	var username, password, identity, prefix, suffix string
	// ClientSideCredential is defined, override the defaults
	if cc := am.ClientSideCredential; cc != nil {
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
	// get the inner auth
	it, err := am.preferredInnerAuthType()
	if err != nil {
		return nil, errors.New("no preferred inner authentication found")
	}

	base.AnonIdentity = identity

	// Configure the credentials and associated metadata
	c := network.Credentials{
		Username: username,
		Prefix:   prefix,
		Suffix:   suffix,
		Password: password,
	}

	return &network.NonTLS{
		Base:        base,
		Credentials: c,
		MethodType:  method.Type(am.EAPMethod.Type),
		InnerAuth:   it,
	}, nil
}

// Network gets a network for an authentication method, the SSID and MinRSN are strings that are based to the network
func (am *AuthenticationMethod) Network(ssid string, minrsn string, pinfo network.ProviderInfo) (network.Network, error) {
	// We check if the eap method is valid
	if am.EAPMethod == nil || !method.IsValid(am.EAPMethod.Type) {
		return nil, errors.New("no EAP method")
	}
	mt := am.EAPMethod.Type

	// Get the server side credentials
	ss := am.ServerSideCredential
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
		Certs:        *CA,
		ProviderInfo: pinfo,
		SSID:         ssid,
		MinRSN:       minrsn,
		ServerIDs:    sid,
	}

	// If TLS we need to construct different arguments than when we have Non TLS
	if method.Type(mt) == method.TLS {
		return am.TLSNetwork(base)
	}
	return am.NonTLSNetwork(base)
}

// Logo returns the logo for the provider info elements
// If the logo is an unexpected type: nil, no PNG, no base64, we return an empty string and an error
// TODO: Note that right now we assume that every logo is a base64 and PNG, we need to determine if this is the only possible type that will be returned
func (pi *ProviderInfoElements) Logo() (string, error) {
	if pi.ProviderLogo == nil {
		return "", errors.New("no provider logo found")
	}
	data := *pi.ProviderLogo
	if data.MimeAttr != "image/png" {
		return "", errors.New("logo is not a PNG image")
	}
	if data.EncodingAttr != "base64" {
		return "", errors.New("image is not base64")
	}
	return data.Value, nil
}

// LocalizedInteractiveValue gets the first non-nil value from the localized interactive slice
// if no value is available, it returns an error
func LocalizedInteractiveValue(slice []*LocalizedInteractive) (string, error) {
	if len(slice) == 0 {
		return "", errors.New("no interactive localized value available")
	}
	for _, v := range slice {
		if v == nil {
			continue
		}
		// TODO: What is LangAttr used for?
		return v.Value, nil
	}
	return "", errors.New("all interactive localized values are nil")
}

// LocalizedNonInteractiveValue gets the first non-nil value from the localized non interactive slice
// if no value is available, it returns an error
func LocalizedNonInteractiveValue(slice []*LocalizedNonInteractive) (string, error) {
	if len(slice) == 0 {
		return "", errors.New("no non interactive localized value available")
	}
	for _, v := range slice {
		if v == nil {
			continue
		}
		// TODO: What is LangAttr used for?
		return v.Value, nil
	}
	return "", errors.New("all non interactive localized values are nil")
}

// PInfo gets the ProviderInfo element from the EAP identity provider
// If it cannot find certain values this will fallback to the default of the type, e.g. an empty string
// TODO: Log errors here
func (p *EAPIdentityProvider) PInfo() network.ProviderInfo {
	var pinfo network.ProviderInfo
	var help network.Help
	pi := p.ProviderInfo
	if pi != nil {
		pinfo.Name, _ = LocalizedNonInteractiveValue(pi.DisplayName)
		pinfo.Description, _ = LocalizedNonInteractiveValue(pi.Description)
		pinfo.Logo, _ = p.ProviderInfo.Logo()
		pinfo.Terms, _ = LocalizedNonInteractiveValue(pi.TermsOfUse)
		desk := p.ProviderInfo.Helpdesk
		if desk != nil {
			help.Email, _ = LocalizedInteractiveValue(desk.EmailAddress)
			help.Web, _ = LocalizedNonInteractiveValue(desk.WebAddress)
			help.Phone, _ = LocalizedInteractiveValue(desk.Phone)
		}
		pinfo.Helpdesk = help
	}
	return pinfo
}

// Network creates a TLS or NON-TLS secured network from the EAP config
// This network can then afterwards be imported into NetworkManager using the `nm` package
func (eap *EAPIdentityProviderList) Network() (network.Network, error) {
	// Get the provider section
	p := eap.EAPIdentityProvider
	if p == nil {
		return nil, errors.New("identity provider section couldn't be found")
	}
	methods, err := p.AuthMethods()
	if err != nil {
		slog.Debug("Error getting AuthMethods", "error", err)
		return nil, err
	}
	ssid, minrsn, err := p.SSIDSettings()
	if err != nil {
		slog.Debug("Error getting SSIDSettings", "error", err)
		return nil, err
	}
	pinfo := p.PInfo()
	for _, m := range methods {
		n, err := m.Network(ssid, minrsn, pinfo)
		if err != nil {
			slog.Debug("Error getting ProviderInfo", "error", err)
			continue
		}
		return n, nil
	}
	return nil, errors.New("no viable network settings found in EAP config")
}
