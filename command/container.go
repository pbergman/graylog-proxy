package command

import (
	"os"

	"github.com/pbergman/app"
	"github.com/pbergman/logger"
	"github.com/pbergman/logger/handlers"
	"github.com/spf13/pflag"
)

type Container struct {
	app     *app.App
	flags   *pflag.FlagSet
	logger  logger.LoggerInterface
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
	if set != c.flags {
		set.AddFlagSet(c.flags)
	} else {
		set.BoolP("quiet", "q", false, "")
		set.StringP("cwd", "c", os.Getenv("PWD"), "")
		set.BoolSliceP("verbose", "v", []bool{}, "")
		set.Lookup("verbose").NoOptDefVal = "true"
		set.Lookup("quiet").NoOptDefVal = "true"
	}
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

func (c *Container) GetLogger() logger.LoggerInterface {
	if c.logger == nil {
		var handler []logger.HandlerInterface
		if ok, _ := c.flags.GetBool("quiet"); ok {
			handler = append(handler, handlers.NewNoOpHandler(logger.DEBUG, false))
		} else {
			switch c.GetVerboseLevel() {
			case 0: // normal
				handler = append(handler, handlers.NewWriterHandler(os.Stdout, logger.WARNING))
			case 1: // verbose
				handler = append(handler, handlers.NewWriterHandler(os.Stdout, logger.NOTICE))
			case 2: // very verbose
				handler = append(handler, handlers.NewWriterHandler(os.Stdout, logger.INFO))
			case 3: // debug
				handler = append(handler, handlers.NewWriterHandler(os.Stdout, logger.DEBUG))

			}
		}
		c.logger = logger.NewLogger("main", handler...)
	}

	if curr := c.GetCurrent(); curr != nil {
		return c.logger.(*logger.Logger).Get(curr.GetName())
	}

	return c.logger
}
