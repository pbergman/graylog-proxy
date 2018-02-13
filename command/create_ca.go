package command

import (
	"fmt"
	"os"
	"time"

	"github.com/pbergman/app"
	"github.com/pbergman/graylog-proxy/x509"
	"github.com/spf13/pflag"
)

func NewCreateCaCommand() app.CommandInterface {
	return &CreateCaCommand{
		create{
			app.Command{
				Flags: new(pflag.FlagSet),
				Name:  "create:ca",
				Usage: "[options] [--] (DN)",
				Short: "Create CA Root certificate and private key",
				Long: `This command will create the root CA certificate and private key that can be used for creating the server and client
certificates .Those certificate can be used in the input config as the "client auth trusted cert" for authenticating
clients and securing the message transport from proxy-node to the graylog server.

Be careful with the force because this will invalidate all already created keys and certificates.

Arguments:
    DN                      The subject used for the certificate as a string (see "help dn")

Options:
    --quiet                 Disable the application output
    --verbose (-v,-vv,-vvv) Increase the verbosity of application output
    --cwd (-c)              Set the current working directory (default '{{ .Env "PWD" }})
    --bits (-b)             The bits size used for creating the private key (default 2048)
    --force (-f)            Overwrite the files if exists (this will invalidate all allready created certificates and keys)
    --not-after (-e)        The the expire time (default 5.256e+6m (10 years))
    --out-pem               The file name for private key (default ./CA_Root.pem)
    --out-crt               The file name for certificate (default ./CA_Root.crt)
    --in-pem                Will use given file as private key and not create a new one. The key should be a PKCS#8 format.

Example:
    {{ exec_bin }} create:ca 'CN=GrayLog Server\, Example,C=Netherlands,L=NL'

`,
			},
		},
	}
}

func (c CreateCaCommand) Init(a *app.App) error {
	if err := c.create.Init(a); err != nil {
		return err
	}
	c.Flags.(*pflag.FlagSet).DurationP("not-after", "e", 5.256e+6*time.Minute, "")
	c.Flags.(*pflag.FlagSet).String("out-pem", "CA_Root.pem", "")
	c.Flags.(*pflag.FlagSet).String("out-crt", "CA_Root.crt", "")
	c.Flags.(*pflag.FlagSet).String("in-pem", "", "")
	return nil
}

type CreateCaCommand struct {
	create
}

func (c CreateCaCommand) Run(args []string, app *app.App) error {
	if s := len(args); s != 1 {
		return fmt.Errorf("invalid arguments, expected 1 got %d", s)
	}
	output := app.Container.(*Container).GetLogger()
	outCrt := c.getFileFromFlag("out-crt")
	outPem := c.getFileFromFlag("out-pem")
	key, err := c.getKey(output, c.getFileFromFlag("in-pem"), outPem)
	if err != nil {
		return err
	}
	crtFile, err := os.OpenFile(outCrt, c.getFlags(), 0600)
	if err != nil {
		return err
	}
	defer crtFile.Close()
	ttl, err := c.getNotAfter()
	if err != nil {
		return err
	}
	if _, err := x509.NewRootCert(key, c.getSubject(args[0]), ttl, crtFile); err != nil {
		return err
	} else {
		output.Debug(fmt.Sprintf("created certificate '%s'", crtFile.Name()))
	}
	return nil
}
