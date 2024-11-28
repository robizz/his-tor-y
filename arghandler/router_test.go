package arghandler

import (
	"errors"
	"testing"

	"github.com/robizz/his-tor-y/conf"
)

type testCommand struct{}

func (t *testCommand) Parse(conf conf.Config, args []string) error { return nil }
func (t *testCommand) Execute() int                                { return 0 }
func (t *testCommand) Help() string                                { return "help" }

func TestRouter(t *testing.T) {

	r := NewRouter()

	var commandName = "test"

	r.Register(commandName, &testCommand{})

	code := r.Execute(conf.Config{}, []string{"main", commandName})

	if code != 0 {
		t.Fatalf("Expected 0, got: %d", code)
	}

}

func TestRouterErrorOnNoArgs(t *testing.T) {

	r := NewRouter()

	r.Register("test", &testCommand{})

	code := r.Execute(conf.Config{}, []string{})

	if code != 1 {
		t.Fatalf("Expected 1, got: %d", code)
	}

}

func TestRouterErrorOnNotEnoughArgs(t *testing.T) {

	r := NewRouter()

	r.Register("test", &testCommand{})

	code := r.Execute(conf.Config{}, []string{"main"})

	if code != 1 {
		t.Fatalf("Expected 1, got: %d", code)
	}

}

func TestRouterErrorOnCommandNotFound(t *testing.T) {

	r := NewRouter()

	var commandName = "test"

	r.Register(commandName, &testCommand{})

	code := r.Execute(conf.Config{}, []string{"main", "not" + commandName})

	if code != 1 {
		t.Fatalf("Expected 1, got: %d", code)
	}

}

type testErrorParseCommand struct{}

func (t *testErrorParseCommand) Parse(conf conf.Config, args []string) error { return errors.New("") }
func (t *testErrorParseCommand) Execute() int                                { return 0 }
func (t *testErrorParseCommand) Help() string                                { return "help" }

func TestRouterErrorOnCommandParseError(t *testing.T) {

	r := NewRouter()

	var commandName = "test"

	r.Register(commandName, &testErrorParseCommand{})

	code := r.Execute(conf.Config{}, []string{"main", commandName})

	if code != 1 {
		t.Fatalf("Expected 1, got: %d", code)
	}

}
