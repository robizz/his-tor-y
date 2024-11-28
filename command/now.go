package command

import (
	"flag"
	"fmt"

	"github.com/robizz/his-tor-y/business"
	"github.com/robizz/his-tor-y/conf"
)

// Command struct
type Now struct {
	StartDate string
	EndDate   string
	Conf      conf.Config
}

func NewNow() *Now {
	return &Now{}
}

func (n *Now) Parse(conf conf.Config, args []string) error {
	n.Conf = conf

	set := flag.NewFlagSet("now", flag.ContinueOnError)
	set.StringVar(&n.StartDate, "start", "2024-01", "The start month in a range search")
	set.StringVar(&n.EndDate, "end", "2024-03", "The end month in a range search")

	if err := set.Parse(args[2:]); err != nil {
		return err
	}

	return nil
}

// implements command interface in main package
func (n *Now) Execute() int {

	out, err := business.Now(n.Conf.ExitNode.DownloadURLTemplate, n.StartDate, n.EndDate)

	if err != nil {
		fmt.Printf("Error %v\n", err)
		return 1
	}
	// Is command package responsible to print the output? if yes should it return 0 right?
	fmt.Println(out)
	return 0

}

func (n *Now) Help() string {
	return "help?"
}
