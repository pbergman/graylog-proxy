package command

import (
	"crypto/rsa"
	"crypto/x509/pkix"
	"encoding/asn1"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/pbergman/app"
	"github.com/pbergman/graylog-proxy/x509"
	"github.com/pbergman/logger"
	"github.com/spf13/pflag"
)

// create is the base command with shared methods
// that are embedded with every create_* command
type create struct {
	app.Command
}

// Init will set the shared/global flags for all the create commands
func (c *create) Init(a *app.App) error {
	a.Container.(*Container).AddFlags(c.Flags.(*pflag.FlagSet))
	c.Flags.(*pflag.FlagSet).IntP("bits", "b", 2048, "")
	c.Flags.(*pflag.FlagSet).BoolP("force", "f", false, "")
	return nil
}

// getBits returns the bits size used for creating the private key
func (c *create) getBits() int {
	raw := c.Flags.(*pflag.FlagSet).Lookup("bits").Value.String()
	value, _ := strconv.Atoi(raw)
	return value
}

// getSubject will convert a subject string to pkix.Name
// that will be used for creating the certificates
func (c *create) getSubject(s string) pkix.Name {
	dn := new(x509.DN)
	dn.ReadString(s)
	return dn.ToPkixName()
}

// getNotAfter will check return the not-after flag
// value used for creating the certificates
func (c *create) getNotAfter() (time.Time, error) {
	duration, err := c.Flags.(*pflag.FlagSet).GetDuration("not-after")
	if err != nil {
		return time.Time{}, err
	} else {
		return time.Now().Add(duration), nil
	}
}

// getFlags will return the mode bit used for opening files
// default will be create/write/excl so if a file exist the
// open method will return a error that the file exists but
// when the force flag is provided the excl flag will
// swapped for the trunc (which will truncate file on open)
func (c create) getFlags() int {
	flags := os.O_CREATE | os.O_WRONLY | os.O_EXCL
	if ok, _ := c.Flags.(*pflag.FlagSet).GetBool("force"); ok {
		flags |= os.O_TRUNC
		flags ^= os.O_EXCL
	}
	return flags
}

// newKey will create a new private key and saves the file
// as pem encoded RSA key in PKCS#8 format
func (c create) newKey(name string) (*rsa.PrivateKey, error) {
	keyFile, err := os.OpenFile(name, c.getFlags(), 0600)
	if err != nil {
		return nil, err
	}
	defer keyFile.Close()
	key, err := x509.NewPrivateKey(c.getBits(), keyFile)
	if err != nil {
		return nil, err
	}
	return key, nil
}

// getKey will return a private keys that can be one that is provided
// with a input filename or when that is a empty string it will create
// a new private key base on the "out" file location
func (c create) getKey(output *logger.Logger, in, out string) (*rsa.PrivateKey, error) {
	if in != "" {
		if key, err := x509.OpenPrivateKey(in); err != nil {
			switch e := err.(type) {
			case *os.PathError:
				return nil, e
			case asn1.StructuralError:
				return nil, fmt.Errorf("failed to decode private key (not a PKCS#8 format?), %s", err.Error())
			default:
				return nil, e
			}
		} else {
			output.Debug(fmt.Sprintf("using private key '%s'", in))
			return key, nil
		}
	} else {
		if key, err := c.newKey(out); err != nil {
			return nil, err
		} else {
			output.Debug(fmt.Sprintf("created private key '%s'", out))
			return key, nil
		}
	}
}

// getFileFromFlag get al full path location based on set CWD and
// give file flag name
func (c create) getFileFromFlag(n string) string {
	return c.resolvePath(c.Flags.(*pflag.FlagSet).Lookup(n).Value.String())
}

// resolvePath will append cwd value to no absolute file locations
func (c create) resolvePath(file string) string {
	if len(file) > 0 && file[0] != '/' {
		return filepath.Join(c.Flags.(*pflag.FlagSet).Lookup("cwd").Value.String(), file)
	}
	return file
}
