package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/robizz/his-tor-y/arghandler"
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
//
// goroutines for multiple download
// new command
// output management
//
// END TODO

const (
	exitCodeErr       = 1
	exitCodeInterrupt = 2
)

// We do this wrapping to allow all defer()s to run before actually exiting.
// See https://pace.dev/blog/2020/02/12/why-you-shouldnt-use-func-main-in-golang-by-mat-ryer.html
func main() {

	// Which configuration?
	conf := conf.Config{
		ExitNode: conf.ExitNode{DownloadURLTemplate: "https://collector.torproject.org/archive/exit-lists/exit-list-%s.tar.xz"},
	}

	// https://pace.dev/blog/2020/02/17/repond-to-ctrl-c-interrupt-signals-gracefully-with-context-in-golang-by-mat-ryer.html
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	defer func() {
		signal.Stop(signalChan)
		cancel()
	}()

	// The select will block until:
	// - we get a signal
	// - the context is done
	// If we get a signal (case <-signalChan) then we call cancel(); this is how we cancel the context when the user presses Ctrl+C.
	// Execution then falls to where we are just waiting for another signal <-signalChan.
	// If it receives anything on the signalChan channel (a second signal),
	// it calls os.Exit which exits without waiting for the program to naturally finish.
	go func() {
		select {
		case <-signalChan: // first signal, cancel context
			cancel()
		case <-ctx.Done():
		}
		// If the program naturally exits after the first signal but before the second,
		// this goroutine will be killed when the program exits (when the main function returns)
		// so there’s no need to go to any extra effort to clean it up, however dissatisfying that might feel.
		<-signalChan // second signal, hard exit
		os.Exit(exitCodeInterrupt)
	}()

	if err := run(ctx, conf, os.Args, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(exitCodeErr)
	}

}

// run wraps the whole code and returns error codes based n errors or 0
// if everything is ok (terminal output is done by System.out stuff)
// the function needs to be integration test friendly tho, meaning we should be
// able to pass parameters and configuration (structs?)
func run(ctx context.Context, conf conf.Config, args []string, stdout io.Writer) error {

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			// create router and register commands
			r := arghandler.NewRouter()
			r.Register("history", command.NewHistory())

			//execute based on args
			return r.Execute(ctx, conf, args, stdout)
		}
	}

}
