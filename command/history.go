package command

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"

	"github.com/robizz/his-tor-y/arghandler"
	"github.com/robizz/his-tor-y/conf"
	"github.com/robizz/his-tor-y/core"
	"github.com/robizz/his-tor-y/exitnode"
)

// Command struct
type History struct {
	StartDate string
	EndDate   string
	IP        string
	Conf      conf.Config
	Output    string
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
	set.StringVar(&n.Output, "output", "text", "The output format")

	if err := set.Parse(args[2:]); err != nil {
		return err
	}

	return nil
}

// implements command interface in main package
func (n *History) Execute(ctx context.Context, stdout io.Writer) error {
	// What am I supposed to do here?
	// An interface would require me to abstract the flags you send to core to make them general
	// or to do even more complicated stuff like "functional options pattern".. just for the sake of testing..
	// An alternative would be to pass a fake download url as did in core tests
	nodes, err := core.History(ctx, n.Conf.ExitNode.DownloadURLTemplate, n.StartDate, n.EndDate, n.IP)

	if err != nil {
		return fmt.Errorf("execute error: %w", err)
	}
	var out string
	switch n.Output {
	case arghandler.Json.String():
		b, err := json.Marshal(&nodes)
		if err != nil {
			return err
		}
		out = string(b)

	case arghandler.Text.String():
		out = table(nodes)

	default:
		out = table(nodes)
	}

	_, err = fmt.Fprint(stdout, out)
	if err != nil {
		return fmt.Errorf("execute error: %w", err)
	}
	return nil
}

func (n *History) Help() string {
	return "help?"
}

func table(nodes []exitnode.ExitNode) string {
	// nice, but now use a https://pkg.go.dev/text/tabwriter and write a test for it with coverage.
	var sb strings.Builder
	sb.WriteString("ExitNode\tPublished\tLastStatus\tExitAddress\tUpdatedAt\n")
	for _, n := range nodes {
		for _, a := range n.ExitAddresses {
			sb.WriteString(fmt.Sprintf("%s\t%s\t%s\t%s\t%s\n", n.ExitNode, n.Published, n.LastStatus, a.ExitAddress, a.UpdatedAt))
		}
	}
	return sb.String()
}
