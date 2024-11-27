// Package arghandler exposes a facility to register the available commands,
// execute the one selected in cLI args and provide an help showing registered
// commands.
package arghandler

import (
	"fmt"

	"github.com/robizz/his-tor-y/conf"
)

// We declare here the "command interface" because we abide to the rules:
// “Go interfaces generally belong in the package that uses values of the interface type,
// not the package that implements those values.”
type Command interface {
	Parse(conf conf.Config, args []string) error
	Execute() int
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

func (r *Router) Execute(conf conf.Config, args []string) int {

	if len(args) < 2 {
		fmt.Printf("Missing command argument")
		return 1
	}

	c, ok := r.commands[args[1]]
	if !ok {
		fmt.Printf("Error command not found")
		//print help here?
		return 1

	}

	err := c.Parse(conf, args)

	if err != nil {
		fmt.Printf("Error %v\n", err)
		return 1
	}

	return c.Execute()
}
