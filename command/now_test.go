package command

import (
	"testing"

	"github.com/robizz/his-tor-y/conf"
)

func TestParse(t *testing.T) {
	n:=NewNow()
	err:=n.Parse(conf.Config{}, []string{"test","now","2024-01","2024-02"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
}

func TestParseErrorOnParsing(t *testing.T) {
	n:=NewNow()
	err:=n.Parse(conf.Config{}, []string{"test","now","2024-01","true", "ttt"})
	if err != nil {
		t.Fatalf("Error expected to be nil")
	}
}