package command

import (
	"fmt"
	"os"
	"time"

	"github.com/pbergman/app"
	"github.com/pbergman/graylog-proxy/x509"
	"github.com/spf13/pflag"
)

func NewCreateServerCommand() app.CommandInterface {
	return &CreateServerCommand{
		create{
			app.Command{
				Flags: new(pflag.FlagSet),
				Name:  "create:server",
				Usage: "[options] [--] (DN)",
				Short: "Create server certificate and private key",
				Long: `This will create the certificate and private key that can be used for creating a TLS based input.

Arguments:
    DN                      The subject used for the certificate as a string (see "help dn")

Options:
    --quiet                 Disable the application output
    --verbose (-v,-vv,-vvv) Increase the verbosity of application output
    --bits (-b)             The bits size used for creating the private key (default 2048)
    --cwd (-c)              Set the current working directory (default '{{ .Env "PWD" }})
    --force (-f)            Overwrite the files if exists
    --host (-H)             Set the server hostname.
    --not-after (-e)        Set the expire time (default 525600m)
    --in-ca-pem             The file name for the CA private key (default ./CA_Root.pem)
    --in-ca--crt            The file name for the CA certificate (default ./CA_Root.crt)
    --in-pem                The file name used for private key
    --out-pem               The file name for the generated private key (default ./Server.pem)
    --out-crt               The file name for the generated certificate (default ./Server.crt)

Example:
    {{ exec_bin }} create:server 'CN=GrayLog Server\, Example,C=Netherlands,L=NL'

`,
			},
		},
	}
}

type CreateServerCommand struct {
	create
}

func (c CreateServerCommand) Init(a *app.App) error {
	if err := c.create.Init(a); err != nil {
		return err
	}
	c.Flags.(*pflag.FlagSet).StringSliceP("host", "H", nil, "")
	c.Flags.(*pflag.FlagSet).DurationP("not-after", "e", 525600*time.Minute, "")
    c.Flags.(*pflag.FlagSet).String("out-pem", "Server.pem", "")
    c.Flags.(*pflag.FlagSet).String("out-crt", "Server.crt", "")
    c.Flags.(*pflag.FlagSet).String("in-pem", "", "")
    c.Flags.(*pflag.FlagSet).String("in-ca-pem", "CA_Root.pem", "")
    c.Flags.(*pflag.FlagSet).String("in-ca-crt", "CA_Root.crt", "")
	return nil
}

func (c CreateServerCommand) Run(args []string, app *app.App) error {
	if s := len(args); s != 1 {
		return fmt.Errorf("invalid arguments, expected 1 got %d", s)
	}
	output := app.Container.(*Container).GetLogger()
	inCrt := c.getFileFromFlag("in-ca-crt")
	inPem := c.getFileFromFlag("in-ca-pem")
	outCrt := c.getFileFromFlag("out-crt")
	outPem := c.getFileFromFlag("out-pem")
	rootCrt, err := x509.OpenCertificate(inCrt)
	if err != nil {
		return err
	} else {
		output.Debug(fmt.Sprintf("using '%s' for CA crt", inCrt))
	}
	rootPem, err := x509.OpenPrivateKey(inPem)
	if err != nil {
		return err
	} else {
		output.Debug(fmt.Sprintf("using '%s' for CA pem", inPem))
	}
    key, err := c.getKey(output, c.getFileFromFlag("in-pem"), outPem)
    if err != nil {
        return err
    }
	ttl, err := c.getNotAfter()
	if err != nil {
		return err
	}
	crtFile, err := os.OpenFile(outCrt, c.getFlags(), 0600)
	if err != nil {
		return err
	}
	defer crtFile.Close()
	if _, err := x509.NewServerCert(key, rootPem, rootCrt, c.getSubject(args[0]), ttl, crtFile); err != nil {
		return err
	} else {
		output.Debug(fmt.Sprintf("created certificate '%s'", crtFile.Name()))
	}
	return nil
}
