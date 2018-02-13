package x509

import (
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/hex"
	"encoding/json"
	"math/big"
	"net"
	"os"
	"time"
)

// helper function to create a basic cert template with a serial number and other required fields
func newCertTemplate(root *x509.Certificate) *x509.Certificate {
	return &x509.Certificate{
		SerialNumber:          getSerialNumber(root),
		SignatureAlgorithm:    x509.SHA256WithRSA,
		NotBefore:             time.Now().Add(-600).UTC(),
		BasicConstraintsValid: true,
	}
}

// getSerialNumber will generate a unique serial number for certificate
// based on the machine id when available and time from now and root cert
func getSerialNumber(root *x509.Certificate) *big.Int {
	serial := new(big.Int)
	// create a base id from the machine id
	if file, err := os.Open("/etc/machine-id"); err == nil {
		defer file.Close()
		// should corresponds to a 16-byte value, that would be 32 byte hex string.
		// see https://www.freedesktop.org/software/systemd/man/machine-id.html
		buf := make([]byte, 32)
		n, _ := file.Read(buf)
		serial.SetBytes(buf[:n])
	}
	if root != nil {
		return serial.Add(big.NewInt((time.Now().Unix()+int64(1))-root.NotBefore.UnixNano()), serial)
	} else {
		return serial.Add(big.NewInt(time.Now().Unix()), serial)
	}
}

// checkSubject will check some fields and set them when needed
func checkSubject(name *pkix.Name) {
	if name.SerialNumber == "" {
		if buf, err := json.Marshal(name); err == nil {
			hash := sha1.Sum(buf)
			name.SerialNumber = hex.EncodeToString(hash[:])
		}
	}
}

// createSubjectKeyId will create a byte slice that represents
// a (SHA-1 hash value) ASN.1 encoding of a public key.
func createSubjectKeyId(key rsa.PublicKey) ([]byte, error) {
	buf, err := asn1.Marshal(key)
	if err != nil {
		return nil, err
	}
	hasher := sha1.New()
	hasher.Write(buf)
	return hasher.Sum(nil), nil
}

func addHosts(tmpl *x509.Certificate, hosts ...string) {
	for _, h := range hosts {
		if ip := net.ParseIP(h); ip != nil {
			tmpl.IPAddresses = append(tmpl.IPAddresses, ip)
		} else {
			tmpl.DNSNames = append(tmpl.DNSNames, h)
		}
	}
}
