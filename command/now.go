package command

import (
	"flag"
	"fmt"

	"github.com/robizz/his-tor-y/conf"
	"github.com/robizz/his-tor-y/core"
)

// Command struct
type Now struct {
	StartDate string
	EndDate   string
	Conf      conf.Config
}

func NewNow(conf conf.Config, args []string) (*Now, error) {

	// Which command?
	now := &Now{}
	now.Conf = conf

	set := flag.NewFlagSet("now", flag.ContinueOnError)
	set.StringVar(&now.StartDate, "start", "2024-01", "The start month in a range search")
	set.StringVar(&now.EndDate, "end", "2024-03", "The end month in a range search")
	// f2 := flag.NewFlagSet("help", flag.ContinueOnError)
	// loud := f2.Bool("loud", false, "")

	if err := set.Parse(args[2:]); err != nil {
		return nil, err
	}

	return now, nil
}

// implements command interface in main package
func (n *Now) Execute() int {

	out, err := core.Now(n.Conf.ExitNode.DownloadURLTemplate, n.StartDate, n.EndDate)

	if err != nil {
		fmt.Printf("Error %v\n", err)
		return 1
	}
	// Is command package responsible to print the output? if yes should it return 0 ritght?
	fmt.Println(out)
	return 0

}
