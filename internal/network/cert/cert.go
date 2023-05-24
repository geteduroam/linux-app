package cert

import (
	"bytes"
	"errors"
	"encoding/base64"
	"encoding/pem"
	"crypto/x509"
)

// toPEM converts an x509 certificate to a PEM encoded block
func toPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

// isRoot checks if a certificate is a root CA
// by checking if the issuer and subject bytes equal and if the certificate is a CA
func isRoot(cert *x509.Certificate) bool {
	return bytes.Equal(cert.RawIssuer, cert.RawSubject) && cert.IsCA
}

type Certs struct {
	Root *x509.Certificate
	Intermediates []*x509.Certificate
}

// ToPEM convers the certs to PEM by first converting the intermediate certificates
// and then the root certificate
func (c *Certs) ToPEM() (ret []byte) {
	// First intermediate certificates
	for _, ic := range c.Intermediates {
		ret = append(ret, toPEM(ic)...)
	}
	// Then the root certificate
	ret = append(ret, toPEM(c.Root)...)
	return ret
}

// New creates a Certs struct by decoding the data in base64
// Note that Certs is guaranteed to be non-nil when there is no error
func New(data []string) (*Certs, error) {
	var root *x509.Certificate
	var inter []*x509.Certificate
	for _, d := range data {
		b, err := base64.StdEncoding.DecodeString(d)
		if err != nil {
			continue
		}
		cert, err := x509.ParseCertificate(b)
		if err != nil {
			continue
		}
		if isRoot(cert) {
			root = cert
		} else {
			inter = append(inter, cert)
		}
	}
	if root == nil {
		return nil, errors.New("no root CA found")
	}
	return &Certs{
		Root: root,
		Intermediates: inter,
	}, nil
}
