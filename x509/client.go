package x509

import (
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"io"
	"time"
)

func NewClientCert(key, rootKey *rsa.PrivateKey, rootCert *x509.Certificate, subject pkix.Name, ttl time.Time, w io.Writer, host ...string) (*x509.Certificate, error) {
	ski, err := createSubjectKeyId(key.PublicKey)
	if err != nil {
		return nil, err
	}
	checkSubject(&subject)
	tmpl := newCertTemplate(rootCert)
	tmpl.Issuer = rootCert.Subject
	tmpl.Subject = subject
	tmpl.NotAfter = ttl
	tmpl.SubjectKeyId = ski
	tmpl.KeyUsage = x509.KeyUsageDigitalSignature
	tmpl.ExtKeyUsage = []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth}
	addHosts(tmpl, host...)
	buf, err := newCert(w, tmpl, rootCert, &key.PublicKey, rootKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(buf)
}
