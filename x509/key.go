package x509

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"

	"golang.org/x/crypto/ssh/terminal"
)

// NewPrivateKey will create a new private key of the given bits size, If the second argument
// is not nil a pem encoded block of the generated key wil be writer to that writer.
func NewPrivateKey(bits int, w io.Writer) (*rsa.PrivateKey, error) {
	if key, err := rsa.GenerateKey(rand.Reader, bits); err != nil {
		return nil, err
	} else {
		if nil != w {
			if buf, err := rsa2pkcs8(key); err != nil {
				return nil, err
			} else {
				block := &pem.Block{
					Type:  "PRIVATE KEY",
					Bytes: buf,
				}
				return key, pem.Encode(w, block)
			}
		} else {
			return key, nil
		}
	}
}

func rsa2pkcs8(key *rsa.PrivateKey) ([]byte, error) {
	var pkey struct {
		Version             int
		PrivateKeyAlgorithm []asn1.ObjectIdentifier
		PrivateKey          []byte
	}
	pkey.Version = 0
	pkey.PrivateKeyAlgorithm = make([]asn1.ObjectIdentifier, 1)
	pkey.PrivateKeyAlgorithm[0] = asn1.ObjectIdentifier{1, 2, 840, 113549, 1, 1, 1}
	pkey.PrivateKey = x509.MarshalPKCS1PrivateKey(key)
	return asn1.Marshal(pkey)
}

// OpenPrivateKey reads the source file and return the private key. The source file
// should be contain a pem encoded private key.
func OpenPrivateKey(name string) (*rsa.PrivateKey, error) {
	if buf, err := ioutil.ReadFile(name); err != nil {
		return nil, err
	} else {
		if block, _ := pem.Decode(buf); block != nil {
			raw := block.Bytes
			if x509.IsEncryptedPEMBlock(block) {
				raw, err = x509.DecryptPEMBlock(block, readPassword("Enter passphrase for key "+filepath.Base(name)+": "))
			}
			key, err := x509.ParsePKCS8PrivateKey(raw)
			if err != nil {
				return nil, err
			} else {
				if v, ok := key.(*rsa.PrivateKey); ok {
					return v, nil
				} else {
					return nil, fmt.Errorf("unsuported key format %T", key)
				}
			}
		} else {
			return nil, fmt.Errorf("failed to decode pem block for file '%s'", name)
		}
	}
}

func readPassword(s string) []byte {
	state, _ := terminal.MakeRaw(syscall.Stdout)
	defer terminal.Restore(syscall.Stdout, state)
	term := terminal.NewTerminal(os.Stdout, ">")
	input, _ := term.ReadPassword(s)
	return []byte(input)
}
