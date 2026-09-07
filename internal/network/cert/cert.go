// Package cert implements parsing of X509 certificates
package cert

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/geteduroam/linux-app/internal/network/cert/rehash"
)

// toPEMFile converts an x509 certificate `cert` to a PEM encoded file `p`
func toPEMFile(p string, cert *x509.Certificate) error {
	gfile, err := os.Create(p)
	if err != nil {
		return err
	}
	defer gfile.Close() //nolint:errcheck
	return pem.Encode(gfile, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

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

// ToDir outputs certificates into a base directory
func (c Certificates) ToDir(baseDir string) error {
	caDir := filepath.Join(baseDir, PEMDir)
	// make sure the CA dir exists
	if err := os.MkdirAll(caDir, 0o700); err != nil {
		return err
	}
	hashes := map[uint32][]string{}
	for k, v := range c {
		fp := filepath.Join(caDir, fmt.Sprintf("%d.pem", k))
		if err := toPEMFile(fp, v); err != nil {
			return err
		}
		hash, err := rehash.SubjectNameHash(v)
		if err != nil {
			return fmt.Errorf("error getting subject name hash for cert (path=%q)\n%w", fp, err)
		}
		hashes[hash] = append(hashes[hash], fp)
	}
	return rehash.CreateSymlinks(caDir, hashes)
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
