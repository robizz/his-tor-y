package main

import (
	"bufio"
	"strings"
	"testing"
)

// TestHelloName calls greetings.Hello with a name, checking
// for a valid return value.
func TestUnmarshall(t *testing.T) {

	someString := `@type tordnsel 1.0
Downloaded 2024-01-30 13:02:00
ExitNode FE39F07EBE7870DCE124AB30DF3ABD0700A43F75
Published 2024-01-30 00:10:50
LastStatus 2024-01-30 10:00:00
ExitAddress 185.241.208.231 2024-01-30 10:21:54
ExitAddress 185.241.208.232 2024-01-30 10:21:54
ExitNode 23B49521BDC4588C7CCF3C38E552504118326B66
Published 2024-01-30 05:44:30
LastStatus 2024-01-30 11:00:00
ExitAddress 194.26.192.64 2024-01-30 11:30:06`

	myReader := strings.NewReader(someString)
	reader := bufio.NewReader(myReader)
	exitNodes,err:=unmarshall(reader)
	if err !=nil{
		t.Fatalf("%v",err)
	}
	if exitNodes[0].ExitNode != "FE39F07EBE7870DCE124AB30DF3ABD0700A43F75" {
	  t.Fatalf("not expected")
	}
	if exitNodes[1].ExitNode != "23B49521BDC4588C7CCF3C38E552504118326B66" {
		t.Fatalf("not expected")
	  }
}
