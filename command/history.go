package command

import (
	"context"
	"flag"
	"fmt"
	"io"

	"github.com/robizz/his-tor-y/conf"
	"github.com/robizz/his-tor-y/core"
)

// Command struct
type History struct {
	StartDate string
	EndDate   string
	IP        string
	Conf      conf.Config
	// here the command should also support an output writer, that
	// I'm going to need to test commands output and formatting and stuff
}

func NewHistory() *History {
	return &History{}
}

func (n *History) Parse(conf conf.Config, args []string) error {
	n.Conf = conf

	set := flag.NewFlagSet("history", flag.ContinueOnError)
	set.StringVar(&n.StartDate, "start", "2024-01", "The start month in a range search")
	set.StringVar(&n.EndDate, "end", "2024-03", "The end month in a range search")
	set.StringVar(&n.IP, "ip", "192.168.1.1", "The IP to search in the TOR nodes history")

	if err := set.Parse(args[2:]); err != nil {
		return err
	}

	return nil
}

// implements command interface in main package
func (n *History) Execute(ctx context.Context, stdout io.Writer) error {
	// What am I supposed to do here?
	// An interface would require me to abstract the flags you send t core to make them general
	// or to do even more complicated stuff like "functional options pattern".. just for the sake of testing..
	// An alternative would be to pass a fake download url as did in core tests
	// out, err := core.Now(ctx, n.Conf.ExitNode.DownloadURLTemplate, n.StartDate, n.EndDate)
	_, err := core.History(ctx, n.Conf.ExitNode.DownloadURLTemplate, n.StartDate, n.EndDate, n.IP)

	if err != nil {
		return fmt.Errorf("execute error: %w", err)
	}

	// Is command package responsible to print the output? if yes should it return 0 right?
	//fmt.Fprintf(stdout, "%s", out)
	fmt.Print(stdout, "done")
	return nil
}

func (n *History) Help() string {
	return "help?"
}
