// package eap implements an XML eap-config parser compliant with the XML schema found at https://github.com/GEANT/CAT/blob/master/devices/eap_config/eap-metadata.xsd
// A part of this was generated with xgen https://github.com/xuri/xgen
// By hand modified:
// - Use NonEAPAuthNumbers as alias instead of hardcoded int
// - Removed -Properties
package eap

import (
	"encoding/xml"
)

// VendorSpecificExtension ...
type VendorSpecificExtension struct {
	VendorAttr int `xml:"vendor,attr"`
}

// TypeSpecificExtension ...
type TypeSpecificExtension struct {
}

// EAPMethod ...
type EAPMethod struct {
	Type           int                        `xml:"Type"`
	TypeSpecific   *TypeSpecificExtension     `xml:"TypeSpecific"`
	VendorSpecific []*VendorSpecificExtension `xml:"VendorSpecific"`
}

// NonEAPAuthNumbers is MSCHAPv2
type NonEAPAuthNumbers int

// IEEE80211RSNProtocols is CTR with CBC-MAC Protocol (if used, only crypto setting
//                         "WPA2/AES" and possible future protos are acceptable).
type IEEE80211RSNProtocols string

// NonEAPAuthMethod ...
type NonEAPAuthMethod struct {
	Type           NonEAPAuthNumbers                        `xml:"Type"`
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
//                 support all types of credentials in the list below. While the
//                 Schema allows to put all kinds of credential information inside
//                 every AuthenticationMethod, even where the information is not
//                 applicable, tags which are not applicable for an authentication
//                 EAP or non-EAP type
//                    SHOULD NOT be included in the corresponding instance of
//                      AuthenticationMethod or InnerAuthenticationMethod when
//                      producing the XML file, and
//                    MUST be ignored by the entity consuming the XML file if
//                      present in the XML file.
type ClientCredentialVariants struct {
	AllowsaveAttr             bool        `xml:"allow_save,attr,omitempty"`
	OuterIdentity             string      `xml:"OuterIdentity"`
	InnerIdentityPrefix       string      `xml:"InnerIdentityPrefix"`
	InnerIdentitySuffix       string      `xml:"InnerIdentitySuffix"`
	InnerIdentityHint         bool        `xml:"InnerIdentityHint"`
	UserName                  string      `xml:"UserName"`
	Password                  string      `xml:"Password"`
	ClientCertificate         *CertData   `xml:"ClientCertificate"`
	IntermediateCACertificate []*CertData `xml:"IntermediateCACertificate"`
	Passphrase                string      `xml:"Passphrase"`
	PAC                       string      `xml:"PAC"`
	ProvisionPAC              bool        `xml:"ProvisionPAC"`
}

// ServerCredentialVariants is Not all EAP types and non-EAP authentication methods need or
//                 support all types of credentials in the list below. While the
//                 Schema allows to put all kinds of credential information inside
//                 every AuthenticationMethod, even where the information is not
//                 applicable, tags which are not applicable for an authentication
//                 EAP or non-EAP type
//                    SHOULD NOT be included in the corresponding instance of
//                      AuthenticationMethod or InnerAuthenticationMethod when
//                      producing the XML file, and
//                    MUST be ignored by the entity consuming the XML file if
//                      present in the XML file.
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
//                 It does e.g. not make sense to have multiple SSIDs in one
//                 IEEE80211 field because the condition would never
//                 match. To specify multiple ORed network properties, use multiple
//                 IEEE80211 instances.
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
	AuthenticationMethod *AuthenticationMethod `xml:"AuthenticationMethod"`
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
