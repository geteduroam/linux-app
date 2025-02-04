package cert

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
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

type Chain struct {
	Roots         []*x509.Certificate
	Intermediates []*x509.Certificate
}

type Certs struct {
	chains map[string]*Chain
}

// ToPEM converts the certs to PEM by first converting the intermediate certificates
// and then the root certificate
func (c *Certs) ToPEM() (ret []byte) {
	// do all the chains one by one
	for _, chain := range c.chains {
		// First intermediate certificates
		for _, ic := range chain.Intermediates {
			ret = append(ret, toPEM(ic)...)
		}
		// Then the root certificates
		for _, root := range chain.Roots {
			ret = append(ret, toPEM(root)...)
		}
	}
	return ret
}

// New creates a Certs struct by decoding the data in base64
// Note that Certs is guaranteed to be non-nil when there is no error
func New(data []string) (*Certs, error) {
	chains := make(map[string]*Chain)
	loopChains := func(wantRoot bool) {
		for _, d := range data {
			b, err := base64.StdEncoding.DecodeString(d)
			if err != nil {
				continue
			}
			cert, err := x509.ParseCertificate(b)
			if err != nil {
				continue
			}
			isRoot := isRoot(cert)
			v, ok := chains[cert.Issuer.String()]
			if wantRoot && isRoot {
				if ok {
					v.Roots = append(v.Roots, cert)
				} else {
					chains[cert.Issuer.String()] = &Chain{Roots: []*x509.Certificate{cert}}
				}
			} else if !wantRoot && !isRoot && ok {
				v.Intermediates = append(v.Intermediates, cert)
			}
		}
	}
	// go through all root CAs
	loopChains(true)
	if len(chains) == 0 {
		return nil, errors.New("no root CA found")
	}
	// then get the intermediate CAs
	loopChains(false)
	return &Certs{
		chains: chains,
	}, nil
}
