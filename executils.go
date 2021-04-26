package executils

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/pkg/errors"
)

type Option struct {
	Cmd     *exec.Cmd
	Verbose bool
}

type OptionFns func(*Option)

func WithVerbose() OptionFns {
	return func(c *Option) {
		c.Verbose = true
	}
}

func WithPath(path string) OptionFns {
	return func(c *Option) {
		c.Cmd.Path = path
	}
}

func WithDir(dir string) OptionFns {
	return func(c *Option) {
		filePath, _ := filepath.Abs(dir)
		c.Cmd.Dir = filePath
	}
}

func WithStdOut(stdOut io.Writer) OptionFns {
	return func(c *Option) {
		c.Cmd.Stdout = stdOut
	}
}

func WithStdErr(stdErr io.Writer) OptionFns {
	return func(c *Option) {
		c.Cmd.Stderr = stdErr
	}
}

func WithStdOutOrErr(stdOutOrErr io.Writer) OptionFns {
	return func(c *Option) {
		c.Cmd.Stderr = stdOutOrErr
		c.Cmd.Stdout = stdOutOrErr
	}
}

func WithEnv(lines ...string) OptionFns {
	return func(c *Option) {
		for _, env := range lines {
			c.Cmd.Env = append(c.Cmd.Env, env)
		}
	}
}

func WithArgs(args ...string) OptionFns {
	return func(c *Option) {
		c.Cmd.Args = append([]string{c.Cmd.String()}, args...)
	}
}

func Run(cmd string, options ...OptionFns) error {
	c := exec.Command(cmd)
	c.Env = os.Environ()

	cmdOptions := &Option{
		Cmd: c,
	}

	for _, opt := range options {
		opt(cmdOptions)
	}

	cmd = os.ExpandEnv(cmd)

	for i := range c.Args {
		c.Args[i] = os.ExpandEnv(c.Args[i])
	}

	if cmdOptions.Verbose {
		fmt.Fprintf(c.Stdout, "Exec: %s\n", strings.Join(c.Args, " "))
	}

	if err := c.Run(); err != nil {
		return errors.WithMessagef(err, "Exit code %d", ExitStatus(err))
	}

	return nil
}

type exitStatus interface {
	ExitStatus() int
}

// ExitStatus returns the exit status of the error if it is an exec.ExitError
// or if it implements ExitStatus() int.
// 0 if it is nil or 1 if it is a different error.
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(exitStatus); ok {
		return e.ExitStatus()
	}
	if e, ok := err.(*exec.ExitError); ok {
		if ex, ok := e.Sys().(exitStatus); ok {
			return ex.ExitStatus()
		}
	}
	return 1
}
