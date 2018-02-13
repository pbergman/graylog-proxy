package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"time"
)

// NewRootCert will generate a new CA cert that will be used to sign and authenticate client and servers,
// a pem encoded block will be written to the given writer (is it is not nil).
func NewRootCert(key *rsa.PrivateKey, subject pkix.Name, ttl time.Time, w io.Writer) (*x509.Certificate, error) {

	ski, err := createSubjectKeyId(key.PublicKey)

	if err != nil {
		return nil, err
	}

	checkSubject(&subject)

	tmpl := newCertTemplate(nil)
	tmpl.IsCA = true
	tmpl.Subject = subject
	tmpl.NotAfter = ttl
	tmpl.SubjectKeyId = ski
	tmpl.KeyUsage = x509.KeyUsageCertSign | x509.KeyUsageDigitalSignature
	tmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth, x509.ExtKeyUsageClientAuth}

	buf, err := newCert(w, tmpl, tmpl, &key.PublicKey, key)

	if err != nil {
		return nil, err
	}

	return x509.ParseCertificate(buf)
}
