// Package arghandler exposes a facility to register the available commands,
// execute the one selected in cLI args and provide an help showing registered
// commands.
package arghandler

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/robizz/his-tor-y/conf"
)

// We declare here the "command interface" because we abide to the rules:
// “Go interfaces generally belong in the package that uses values of the interface type,
// not the package that implements those values.”
type Command interface {
	Parse(conf.Config, []string) error
	Execute(context.Context, io.Writer) error
	Help() string
}

type Router struct {
	commands map[string]Command
}

func NewRouter() *Router {
	return &Router{
		commands: make(map[string]Command),
	}
}

func (r *Router) Register(name string, c Command) {
	r.commands[name] = c
}

// what about the help here it should run after everything is registered right?

func (r *Router) Execute(ctx context.Context, conf conf.Config, args []string, stdout io.Writer) error {

	if len(args) < 2 {
		return errors.New("missing command")
	}

	c, ok := r.commands[args[1]]
	if !ok {
		//print help here?
		return errors.New("command not found")

	}

	err := c.Parse(conf, args)

	if err != nil {
		return fmt.Errorf("parse error: %w", err)
	}

	return c.Execute(ctx, stdout)
}
