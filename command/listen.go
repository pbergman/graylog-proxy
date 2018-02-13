package command

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/pbergman/app"
	"github.com/pbergman/graylog-proxy/net"
	"github.com/spf13/pflag"
)

func NewListenCommand() app.CommandInterface {
	return &ListenCommand{
		app.Command{
			Flags: new(pflag.FlagSet),
			Name:  "listen",
			Usage: "[options] [--] (LOCAL_ADDRESS) [REMOTE_ADDRESS]",
			Short: "Start message forwarder",
			Long: `This listen to the given LOCAL_ADDRESS and forward all incoming message to the REMOTE_ADDRESS. The LOCAL_ADDRESS
and REMOTE_ADDRESS should be in the format of scheme://address and where scheme for LOCAL_ADDRESS is a connectionless
protocol like unixgram, udp or ip and  REMOTE_ADDRESS scheme should be one of tcp, tcp+ssl, http or https (see help host).

When using the "print" flag the REMOTE_ADDRESS argument becomes optional and will only dump the incoming messages when
the REMOTE_ADDRESS is not provided.

    --quiet                 Disable the application output
    --verbose (-v,-vv,-vvv) Increase the verbosity of application output
	--cwd (-c)              Set the current working directory (default '{{ .Env "PWD" }})
    --pem                   The file name for the client private key (default ./Client.pem)
    --crt                   The file name for the client certificate (default ./Client.crt)
    --ca                    The file name for the CA certificate (default ./CA_Root.crt)
    --new-line              Use new line delimiter instead of a null byte
	--workers (-w)          Set the max concurrent workers for handling incoming messages (default 10)
	--print                 Print the message as the are going to be send
	--no-client-auth        Will nog load certificates hen using a secure scheme

Example:
    {{ exec_bin }} listen 127.0.0.1:12201 example.logger.com:12201
`,
		},
	}
}

type ListenCommand struct {
	app.Command
}

func (c *ListenCommand) Init(a *app.App) error {
	a.Container.(*Container).AddFlags(c.Flags.(*pflag.FlagSet))
	c.Flags.(*pflag.FlagSet).String("pem", "Client.pem", "")
	c.Flags.(*pflag.FlagSet).String("crt", "Client.crt", "")
	c.Flags.(*pflag.FlagSet).String("ca", "CA_Root.crt", "")
	c.Flags.(*pflag.FlagSet).Bool("new-line", false, "")
	c.Flags.(*pflag.FlagSet).Lookup("new-line").NoOptDefVal = "true"
	c.Flags.(*pflag.FlagSet).IntP("workers", "w", 10, "")
	c.Flags.(*pflag.FlagSet).BoolP("print", "p", false, "")
	c.Flags.(*pflag.FlagSet).Lookup("print").NoOptDefVal = "true"
	c.Flags.(*pflag.FlagSet).Bool("no-client-auth", false, "")
	c.Flags.(*pflag.FlagSet).Lookup("no-client-auth").NoOptDefVal = "true"
	return nil
}

func (c ListenCommand) getIntVar(s string) int {
	r, _ := c.Flags.(*pflag.FlagSet).GetInt(s)
	return r
}

func (c ListenCommand) getBoolVar(s string) bool {
	r, _ := c.Flags.(*pflag.FlagSet).GetBool(s)
	return r
}

func (c *ListenCommand) print() bool {
	v, _ := c.Flags.(*pflag.FlagSet).GetBool("print")
	return v
}

func (c ListenCommand) getFileFromFlag(n string) string {
	file := c.Flags.(*pflag.FlagSet).Lookup(n).Value.String()
	if file[0] != '/' {
		file = filepath.Join(c.Flags.(*pflag.FlagSet).Lookup("cwd").Value.String(), file)
	}
	return file
}

func (c *ListenCommand) Run(args []string, app *app.App) error {

	isPrint := c.print()

	if isPrint && (len(args) != 1 && len(args) != 2) {
		return fmt.Errorf("invalid arguments, expected 1 got %d", len(args))
	}

	if !isPrint && len(args) != 2 {
		return fmt.Errorf("invalid arguments, expected 2 got %d", len(args))
	}

	var local, remote string
	var conn net.ConnPoolInterface

	local = args[0]

	if !isPrint || len(args) >= 2 {
		remote = args[1]
	}

	if !isPrint && remote == "" {
		return errors.New("missing remote")
	}

	if strings.Index(local, "://") == -1 {
		local = "udp://" + local
	}

	if remote != "" && strings.Index(remote, "://") == -1 {
		remote = "tcp+ssl://" + remote
	}

	logger := app.Container.(*Container).GetLogger()

	if !isPrint || remote != "" {
		logger.Debug(fmt.Sprintf("starting forward %s → %s", local, remote))
	}

	listener, err := net.NewListener(local, logger)

	if err != nil {
		return err
	}

	defer listener.Close()

	if remote != "" {
		host := net.NewGraylogHost(remote)

		if host == nil {
			return fmt.Errorf("invalid hostname provided, see \"help host\"")
		}

		conn, err = net.NewConnPool(
			c.getBoolVar("no-client-auth"),
			c.getIntVar("tries"),
			host,
			c.Flags.(*pflag.FlagSet).Lookup("ca").Value.String(),
			c.Flags.(*pflag.FlagSet).Lookup("crt").Value.String(),
			c.Flags.(*pflag.FlagSet).Lookup("pem").Value.String(),
			app.Container.(*Container).GetLogger(),
		)

		if err != nil {
			return err
		}
		defer conn.Close()

		workers, _ := c.Flags.(*pflag.FlagSet).GetInt("workers")
		logger.Debug(fmt.Sprintf("starting conn %d workers", workers))
		conn.Start(workers)
	}

	go listener.Listen()

	for ret := range listener.Done {
		switch val := ret.(type) {
		case *net.FatalError:
			return val
		case error:
			logger.Error(val)
		case []byte:
			id, message := val[:8], val[8:]
			if isPrint {
				fmt.Printf("\n#### %X ####\n%s\n##########################\n\n", id, message)
			}
			if conn != nil {
				if c.getBoolVar("new-line") {
					conn.Push(append(message, '\n'), id)
				} else {
					conn.Push(append(message, byte(0)), id)
				}
			}
		}
	}

	return nil
}
