package command

import (
	"github.com/pbergman/app"
)

func NewHostCommand() app.CommandInterface {
	return &app.Command{
		Name:  "host",
		Short: "Information about the host argument",
		Long: `
The host argument should be: (scheme://remote[:ip]) where scheme should be a valid scheme and remote representing the
remote host and ip separated by a colon.

This application support 2 input types ("GELF TCP" and "GELF HTTP") as input to forward the message to. Valid schemes
for tcp are tcp, tcp4, tcp6, tcp+tls, tcp4+tls, tcp6+tls where tcp(4|6) represents an unencrypted plain connection and
tcp[4|6]+tls represents a tls over tcp connection.

example:

    tcp+tls://127.0.0.1:12201   # to connect to remote 127.0.0.1 using tls over tcp on port 12201
    http://127.0.0.1/gelf       # for a http input.
`,
	}
}
