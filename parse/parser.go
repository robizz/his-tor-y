package parse

import (
	"bufio"
	"fmt"
	"io"
	"strings"
	"time"
)

type ExitNode struct {
	ExitNode      string        `json:"ExitNode"`
	Published     time.Time     `json:"Published"`
	LastStatus    time.Time     `json:"LastStatus"`
	ExitAddresses []ExitAddress `json:"ExitAddresses"`
}

type ExitAddress struct {
	ExitAddress string    `json:"ExitAddress"`
	UpdatedAt   time.Time `json:"UpdatedAt"`
}

func Unmarshall(r *bufio.Reader) ([]ExitNode, error) {
	exitNodes := []ExitNode{}
	var exitNode ExitNode
	for {
		// Reading a line, lines are short so we don't worry about getting truncated/prefixes.
		line, _, err := r.ReadLine()
		if err != nil {
			if err == io.EOF {
				if exitNode.ExitNode != "" {
					exitNodes = append(exitNodes, exitNode)
				}
				break
			}
			return nil, err
		}

		// here starts marshaller logic
		split := strings.Split(string(line), " ")
		key := split[0]
		// here I'm removing the key from the line to get a number of values (could be 1, 2, or 3 values depending on the entry).
		value := strings.Replace(string(line), key+" ", "", 1)
		// Here I'm splitting the value part, and I'm sure that at least every line type is going to have at least one value.
		values := strings.Split(value, " ")

		switch key {
		//headers, so we skip
		case "@type":
		case "Downloaded":
			continue
		case "ExitNode":
			// If the current ExitNode is not empty, we append it in the list and we move on with a new one.
			if exitNode.ExitNode != "" {
				exitNodes = append(exitNodes, exitNode)
			}
			// Time sto start filling a new ExitNode struct
			exitNode = ExitNode{}
			exitNode.ExitNode = values[0]
		case "Published":
			u, err := time.Parse(time.RFC3339, values[0]+"T"+values[1]+"Z")
			if err != nil {
				return nil, fmt.Errorf("field Published date parse error: %w", err)
			}
			exitNode.Published = u
		case "LastStatus":
			u, err := time.Parse(time.RFC3339, values[0]+"T"+values[1]+"Z")
			if err != nil {
				return nil, fmt.Errorf("field LastStatus date parse error: %w", err)
			}
			exitNode.LastStatus = u
		case "ExitAddress":
			u, err := time.Parse(time.RFC3339, values[1]+"T"+values[2]+"Z")
			if err != nil {
				return nil, fmt.Errorf("field ExitAddress date parse error: %w", err)
			}
			e := ExitAddress{
				ExitAddress: values[0],
				UpdatedAt:   u,
			}
			exitNode.ExitAddresses = append(exitNode.ExitAddresses, e)
		default:
			// skip
			// fmt.Println(key)
		}
	}
	return exitNodes, nil
}
