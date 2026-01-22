package cert

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"log/slog"
)

// toPEM converts an x509 certificate to a PEM encoded block
func toPEM(cert *x509.Certificate) []byte {
	return pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: cert.Raw})
}

// New creates certs by decoding each certificate and continuing on invalid certificates
// It returns PEM encoded data
func New(data []string) ([]byte, error) {
	var ret []byte
	for _, v := range data {
		b, err := base64.StdEncoding.DecodeString(v)
		if err != nil {
			slog.Error("failed parsing certificate", "error", err)
			continue
		}
		cert, err := x509.ParseCertificate(b)
		if err != nil {
			slog.Error("failed parsing certificate", "error", err)
			continue
		}
		ret = append(ret, toPEM(cert)...)
	}

	if len(data) == 0 {
		return nil, errors.New("no viable certificates found")
	}
	return ret, nil
}
