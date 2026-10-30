// Package cert implements parsing of X509 certificates
package cert

import (
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func removeIfExists(p string) error {
	if _, err := os.Stat(p); err == nil {
		if err := os.RemoveAll(p); err != nil {
			return err
		}
	}
	return nil
}

// Certificates is a list of x509 Certificates
type Certificates []*x509.Certificate

// ToPEM outputs PEM encoded blocks for all the certificates
func (c Certificates) ToPEM() []byte {
	var ret []byte
	for _, v := range c {
		ret = append(ret, toPEM(v)...)
	}
	return ret
}

const (
	// PEMFile is the file where certificate data is stored
	PEMFile = "ca-cert.pem"

	// PEMDir is the directory where the PEM certificates are stored
	PEMDir = "ca"
)

// Cleanup removes old certificate data
func Cleanup(baseDir string) error {
	caDir := filepath.Join(baseDir, PEMDir)
	// remove the previous CA directory
	if err := removeIfExists(caDir); err != nil {
		return err
	}
	// remove a ca-cert.pem in the base dir
	if err := removeIfExists(filepath.Join(baseDir, PEMFile)); err != nil {
		return err
	}
	return nil
}

// New creates certs by decoding each certificate and continuing on invalid certificates
// It returns PEM encoded data
func New(data []string) (Certificates, error) {
	ret := make(Certificates, len(data))
	for i, v := range data {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			return nil, fmt.Errorf("failed decoding base64 for certificate: %w", err)
		}
		cert, err := x509.ParseCertificate(b)
		if err != nil {
			return nil, fmt.Errorf("failed parsing certificate: %w", err)
		}
		ret[i] = cert
	}

	if len(data) == 0 {
		return nil, errors.New("no viable certificates found")
	}
	return ret, nil
}
