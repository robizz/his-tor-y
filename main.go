package main

import (
	"fmt"
	"os"

	"github.com/robizz/his-tor-y/command"
	"github.com/robizz/his-tor-y/conf"
)

// TODO:
// probably the generate file list should go in the extract package and be a deeper module...
// also the nodes structure make the package parse dumb as a name mmmm
/*
packages proposal:
- command
- download
- extract
- parse
- transform
- output
*/

// related to commands: I need:
// - a way to define a command interface with related inputs and outputs
// - a way to inject specific flags to specific commands
// - a way to define a strategy pattern kinda of approach to know which command to instantiate or launch

// command line options to tune the resolution of the compaction
// when treating multiple days, duplicates management needs to be managed.
// a final cleanup of all text files must be done
// are we sure we want to use pointers for exit nodes? for now we have values, maybe a memory footprint and performance instrumentation with a full year of data would be nice
// When program reaches the desired complexity and tests are in place, apply effective go / practical go / bill kennedy refactoring
// clean comments
// variable names are ugly
// create a cache and allow commands to run in the cache (maybe using a bolt db? an embedded database? an in memory struct?)
// the in memory struct could be also a zipped json or array of zipped items of a struct that you decompress on the fly, perf it would be nice.
// command should be silent to use pipe or output redirect. errors should be on stderr
// errors should be constant errors like dave cheney suggests
// we need an integration test to test the whole flow
// Main functionality is: I give you the list of nodes that were found for the time range with the last update inside the time range.
// another funtionality is "IP History":I give you an IP and a parameter like "days", the tool gives me 0 with formatted list of nodes and dates.
// generate go doc
// END TODO

// We declare here the "command interface" because we abide to the rules:
// “Go interfaces generally belong in the package that uses values of the interface type, not the package that implements those values.”
type Command interface {
	Execute() int
}

// We do this wrapping to allow all defer()s to run before actually exiting.
func main() {

	// Which configuration?
	conf := conf.Config{
		ExitNode: conf.ExitNode{DownloadURLTemplate: "https://collector.torproject.org/archive/exit-lists/exit-list-%s.tar.xz"},
	}

	os.Exit(mainReturnWithCode(conf, os.Args))

}

// mainReturnWithCode wraps the whole code and returns error codes based n errors or 0
// if everything is ok (terminal output is done by System.out stuff)
// the function needs to be integration test friendly tho, meaning we should be
// able to pass parameters and configuration (structs?)
func mainReturnWithCode(conf conf.Config, args []string) int {

	var c Command
	var err error
	switch args[1] {
	case "now":
		c, err = command.NewNow(conf, args)

	// command not found
	// I should print the help here
	// but the help has to list available commands with their options tho
	default:
		fmt.Println("ay")
		return 1
	}
	if err != nil {
		fmt.Printf("Error %v\n", err)
		return 1
	}
	return c.Execute()
}
