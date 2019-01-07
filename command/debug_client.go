package command

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"time"

	"github.com/pbergman/app"
	"github.com/pbergman/graylog-proxy/net"
	"github.com/spf13/pflag"
)

func NewDebugClientCommand() app.CommandInterface {
	return &DebugClientCommand{
		create: create{
			app.Command{
				Flags: new(pflag.FlagSet),
				Name:  "debug:client",
				Usage: "[options] [--] (HOST)",
				Short: "Send a GELF message log to the given host",
				Long: `debug:client command will send a message to the given host to debug/test the connection.

Arguments:
    HOST                    the host of the graylog input (see help host)

Options:
    --quiet                 Disable the application output
    --verbose (-v,-vv,-vvv) Increase the verbosity of application output
    --cwd (-c)              Set the current working directory (default '{{ .Env "PWD" }})
    --full-message          Full message that will be used for the GELF payload (default will be a stack trace)
    --short-message         Short message that will be used for the GELF payload (default 'example stack trace')
    --host                  Host that will be used for the GELF payload (default to the hostname)
    --tries (-t)            The amount of re-tries before we give up delivering the message
    --level                 Level for the GELF payload (default 1)
    --dump                  Write a hexdump of the message that's going to be send to the server.
    --new-line              Use new line delimiter instead of a null byte
    --no-client-auth        Will nog load certificates hen using a secure scheme
    --pem                   The file name for the client private key (default ./Client.pem)
    --crt                   The file name for the client certificate (default ./Client.crt)
    --ca                    The file name for the CA certificate (default ./CA_Root.crt)

Example:
    {{ exec_bin }} debug:client --pem=client.pem --crt=client.crt --ca=root.crt tcp+tls://example.logger.com:12201

`,
			},
		},
	}
}

type DebugClientCommand struct {
	create
}

func (c *DebugClientCommand) Init(a *app.App) error {
	if err := c.create.Init(a); err != nil {
		return err
	}
	c.Flags.(*pflag.FlagSet).String("pem", "Client.pem", "")
	c.Flags.(*pflag.FlagSet).String("crt", "Client.crt", "")
	c.Flags.(*pflag.FlagSet).String("ca", "CA_Root.crt", "")
	c.Flags.(*pflag.FlagSet).String("full-message", "", "")
	c.Flags.(*pflag.FlagSet).IntP("tries", "t", 5, "")
	c.Flags.(*pflag.FlagSet).String("short-message", "example stack trace", "")
	c.Flags.(*pflag.FlagSet).String("host", "", "")
	c.Flags.(*pflag.FlagSet).Bool("new-line", false, "")
	c.Flags.(*pflag.FlagSet).Bool("dump", false, "")
	c.Flags.(*pflag.FlagSet).Bool("no-client-auth", false, "")
	c.Flags.(*pflag.FlagSet).Int8("level", 1, "")
	c.Flags.(*pflag.FlagSet).Lookup("new-line").NoOptDefVal = "true"
	c.Flags.(*pflag.FlagSet).Lookup("no-client-auth").NoOptDefVal = "true"
	c.Flags.(*pflag.FlagSet).Lookup("dump").NoOptDefVal = "true"
	return nil
}

func (c DebugClientCommand) getIntVar(s string) int {
	r, _ := c.Flags.(*pflag.FlagSet).GetInt(s)
	return r
}

func (c DebugClientCommand) getBoolVar(s string) bool {
	r, _ := c.Flags.(*pflag.FlagSet).GetBool(s)
	return r
}

func (c DebugClientCommand) getMessage() (buf []byte, err error) {
	level, _ := c.Flags.(*pflag.FlagSet).GetInt8("level")
	data := map[string]interface{}{
		"version":       "1.1",
		"host":          c.Flags.(*pflag.FlagSet).Lookup("host").Value.String(),
		"short_message": c.Flags.(*pflag.FlagSet).Lookup("short-message").Value.String(),
		"full_message":  c.Flags.(*pflag.FlagSet).Lookup("full-message").Value.String(),
		"timestamp":     time.Now().Unix(),
		"level":         level,
	}
	if "" == data["host"] {
		data["host"], err = os.Hostname()
		if err != nil {
			return
		}
	}
	if "" == data["full_message"] {
		buffer := make([]byte, 4086)
		n := runtime.Stack(buffer, true)
		data["full_message"] = string(buffer[:n])
	}
	buf, err = json.Marshal(data)
	if err != nil {
		return
	}
	if c.getBoolVar("new-line") {
		return append(buf, '\n'), nil
	} else {
		return append(buf, byte(0)), nil
	}
}

func (c DebugClientCommand) Run(args []string, app *app.App) error {
	if s := len(args); s != 1 {
		return fmt.Errorf("invalid arguments, expected 1 got %d", s)
	}
	host := net.NewGraylogHost(args[0])
	if host == nil {
		return fmt.Errorf("invalid hostname provided, see \"help host\"")
	}
	pool, err := net.NewConnPool(
		c.getBoolVar("no-client-auth"),
		c.getIntVar("tries"),
		host,
		c.getFileFromFlag("ca"),
		c.getFileFromFlag("crt"),
		c.getFileFromFlag("pem"),
		app.Container.(*Container).GetLogger(),
	)
	if err != nil {
		return err
	}
	defer pool.Close()
	pool.Start(1)
	message, err := c.getMessage()
	if err != nil {
		return err
	}
	if c.getBoolVar("dump") {
		fmt.Print(hex.Dump(message))
	}
	_, err = pool.Write(message)
	if err != nil {
		return err
	}
	return nil
}
