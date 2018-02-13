package x509

import (
	"crypto/rand"
	"crypto/x509"
	"encoding/pem"
	"io"
	"io/ioutil"
)

// createCert wil create and write the cert to given file, if exist it will return os.ErrExist
func newCert(w io.Writer, tmpl, parent *x509.Certificate, pub, priv interface{}) ([]byte, error) {
	if cert, err := x509.CreateCertificate(rand.Reader, tmpl, parent, pub, priv); err != nil {
		return nil, err
	} else {
		if w != nil {
			block := &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: cert,
			}
			return cert, pem.Encode(w, block)
		} else {
			return cert, nil
		}
	}
}

// OpenCertificate opem PEM encoded certificate file and decode data
func OpenCertificate(name string) (*x509.Certificate, error) {
	if buf, err := ioutil.ReadFile(name); err != nil {
		return nil, err
	} else {
		block, _ := pem.Decode(buf)
		return x509.ParseCertificate(block.Bytes)
	}
}
