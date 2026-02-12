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
	defer gfile.Close()
	pem.Encode(gfile, &pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
	return nil
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

// ToDir outputs certificates into a base directory
func (c Certificates) ToDir(baseDir string) error {
	caDir := filepath.Join(baseDir, "ca")
	// remove the previous CA directory
	if err := removeIfExists(caDir); err != nil {
		return err
	}
	// remove a ca-cert.pem in the base dir which is from an old version of the client
	if err := removeIfExists(filepath.Join(baseDir, "ca-cert.pem")); err != nil {
		return err
	}
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
