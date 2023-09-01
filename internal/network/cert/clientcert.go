package cert

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"time"

	"github.com/youmark/pkcs8"
	"software.sslmate.com/src/go-pkcs12"
)

// ClientCert is the client certificate structure
type ClientCert struct {
	// cert is the actual client certificate obtained from the PKCS12 container
	cert *x509.Certificate
	// privateKey is the RSA private key obtained from the PKCS12 container
	privateKey interface{}
}

// NewClientCert creates a new client certificate using the pkcs12 string 'pkcs12s' and passphrase 'pass'
// It returns the client certificate object
func NewClientCert(pkcs12s string, pass string) (*ClientCert, error) {
	rawcc, err := base64.StdEncoding.DecodeString(pkcs12s)
	if err != nil {
		return nil, err
	}
	// decode the PKCS12 container to get the client certificate
	pk, cc, _, err := pkcs12.DecodeChain(rawcc, pass)
	if err != nil {
		return nil, err
	}
	return &ClientCert{
		cert:       cc,
		privateKey: pk,
	}, nil
}

// genb64 creates a cryptographically random bytes slice of 32 bytes
// This byte slice is then encoded to base64
// It returns the byte slice encoded to base64 (or nil if error) and an error if it could not be generated.
func genb64() (pwd string, err error) {
	n := 32
	bs := make([]byte, n)
	if _, err := rand.Read(bs); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(bs), nil
}

// PrivateKeyPEMEnc gets the private key in encrypted PEM format
// It returns the PEM format for the private key, the password protecting it and an error
// The password is automatically generated by generating random bytes using crypto/rand
// ... and then feeding it to base64 encoding
func (cc *ClientCert) PrivateKeyPEMEnc() (pemb []byte, pwd string, err error) {
	pwd, err = genb64()
	if err != nil {
		return nil, "", err
	}
	b, err := pkcs8.MarshalPrivateKey(cc.privateKey, []byte(pwd), nil)
	if err != nil {
		return nil, "", err
	}
	block := &pem.Block{
		Type:  "ENCRYPTED PRIVATE KEY",
		Bytes: b,
	}
	return pem.EncodeToMemory(block), pwd, nil
}

// ToPEM generates the PEM bytes for the client certificate
func (cc *ClientCert) ToPEM() []byte {
	return toPEM(cc.cert)
}

func (cc *ClientCert) Validity() time.Time {
	return cc.cert.NotAfter
}
