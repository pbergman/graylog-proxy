package command

import (
	"os"

	"github.com/pbergman/app"
	"github.com/pbergman/logger"
	"github.com/spf13/pflag"
)

type Container struct {
	app     *app.App
	flags   *pflag.FlagSet
	logger  *logger.Logger
	current app.CommandInterface
}

func (c *Container) SetCurrent(a app.CommandInterface) {
	c.current = a
}

func (c *Container) Env(s string) string {
	return os.Getenv(s)
}

func (c *Container) GetCurrent() app.CommandInterface {
	return c.current
}

func (c *Container) AddFlags(set *pflag.FlagSet) {
	set.BoolP("quiet", "q", false, "")
	set.StringP("cwd", "c", os.Getenv("PWD"), "")
	set.BoolSliceP("verbose", "v", []bool{}, "")
	set.BoolP("help", "h", false, "")
	set.Lookup("help").NoOptDefVal = "true"
	set.Lookup("verbose").NoOptDefVal = "true"
	set.Lookup("quiet").NoOptDefVal = "true"
}

func (c *Container) init() {
	if c.flags == nil {
		c.flags = pflag.NewFlagSet("main", pflag.ContinueOnError)
		c.flags.Usage = func() { c.GetApp().Usage(os.Stderr); os.Exit(2) }
		c.AddFlags(c.flags)
	}
	c.flags.Parse(os.Args[1:])
}

func (c *Container) GetArgs() []string {
	args := c.GetFlags().Args()

	if len(args) <= 0 {
		return args
	}

	// try to get raw args not parsed, this so
	// second time called parse we wont mis the
	// -- args
	for i, c := 1, len(os.Args); i < c; i++ {
		if os.Args[i] == args[0] {
			return os.Args[i:]
		}
	}
	return args
}

func (c *Container) GetFlags() *pflag.FlagSet {
	if c.flags == nil {
		c.init()
	}
	return c.flags
}

func (c *Container) GetApp() *app.App {
	return c.app
}

func (c *Container) SetApp(a *app.App) {
	c.app = a
}

func (c *Container) GetVerboseLevel() int {
	size, _ := c.flags.GetBoolSlice("verbose")
	return len(size)
}

func (c *Container) GetLogger() *logger.Logger {
	if c.logger == nil {
		var handler []logger.HandlerInterface
		if ok, _ := c.flags.GetBool("quiet"); !ok {
			switch c.GetVerboseLevel() {
			case 0: // normal
				handler = append(handler, logger.NewWriterHandler(os.Stdout, logger.LogLevelWarning(), false))
			case 1: // verbose
				handler = append(handler, logger.NewWriterHandler(os.Stdout, logger.LogLevelNotice(), false))
			case 2: // very verbose
				handler = append(handler, logger.NewWriterHandler(os.Stdout, logger.LogLevelInfo(), false))
			default: // debug
				handler = append(handler, logger.NewWriterHandler(os.Stdout, logger.LogLevelDebug(), false))
			}
		}
		c.logger = logger.NewLogger("main", handler...)
	}

	if curr := c.GetCurrent(); curr != nil {
		return c.logger.WithName(curr.GetName())
	}

	return c.logger
}
