package command

import (
	"flag"
	"fmt"

	"github.com/robizz/his-tor-y/business"
	"github.com/robizz/his-tor-y/conf"
)

// Command struct
type NowFlags struct {
	StartDate string
	EndDate   string
}

func Now(conf conf.Config, args []string) int {

	// Which command?
	nowFlags := NowFlags{}
	set := flag.NewFlagSet("now", flag.ContinueOnError)
	set.StringVar(&nowFlags.StartDate, "start", "2024-01", "The start month in a range search")
	set.StringVar(&nowFlags.EndDate, "end", "2024-03", "The end month in a range search")
	// f2 := flag.NewFlagSet("help", flag.ContinueOnError)
	// loud := f2.Bool("loud", false, "")

	if err := set.Parse(args[2:]); err != nil {
		fmt.Printf("Error %v\n", err)
		return 1
	}

	out, err := business.Now(conf.ExitNode.DownloadURLTemplate, nowFlags.StartDate, nowFlags.EndDate)

	if err != nil {
		fmt.Printf("Error %v\n", err)
		return 1
	}
	// Is command package responsible to print the output? if yes should it return 0 ritght?
	fmt.Println(out)
	return 0

}
