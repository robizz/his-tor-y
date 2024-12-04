package arghandler

import (
	"context"
	"errors"
	"io"
	"os"
	"testing"

	"github.com/robizz/his-tor-y/conf"
)

type testCommand struct{}

func (t *testCommand) Parse(conf conf.Config, args []string) error { return nil }
func (t *testCommand) Execute(context.Context, io.Writer) error    { return nil }
func (t *testCommand) Help() string                                { return "help" }

func TestRouter(t *testing.T) {

	r := NewRouter()

	var commandName = "test"

	r.Register(commandName, &testCommand{})

	err := r.Execute(context.Background(), conf.Config{}, []string{"main", commandName}, os.Stdout)

	if err != nil {
		t.Fatalf("Expected nil, got: %v", err)
	}

}

func TestRouterErrorOnNoArgs(t *testing.T) {

	r := NewRouter()

	r.Register("test", &testCommand{})

	code := r.Execute(context.Background(), conf.Config{}, []string{}, os.Stdout)

	if code == nil {
		t.Fatalf("Expected error, got: nil")
	}

}

func TestRouterErrorOnNotEnoughArgs(t *testing.T) {

	r := NewRouter()

	r.Register("test", &testCommand{})

	code := r.Execute(context.Background(), conf.Config{}, []string{"main"}, os.Stdout)

	if code == nil {
		t.Fatalf("Expected error, got: nil")
	}

}

func TestRouterErrorOnCommandNotFound(t *testing.T) {

	r := NewRouter()

	var commandName = "test"

	r.Register(commandName, &testCommand{})

	code := r.Execute(context.Background(), conf.Config{}, []string{"main", "not" + commandName}, os.Stdout)

	if code == nil {
		t.Fatalf("Expected error, got: nil")
	}

}

type testErrorParseCommand struct{}

func (t *testErrorParseCommand) Parse(conf conf.Config, args []string) error { return errors.New("") }
func (t *testErrorParseCommand) Execute(context.Context, io.Writer) error    { return nil }
func (t *testErrorParseCommand) Help() string                                { return "help" }

func TestRouterErrorOnCommandParseError(t *testing.T) {

	r := NewRouter()

	var commandName = "test"

	r.Register(commandName, &testErrorParseCommand{})

	err := r.Execute(context.Background(), conf.Config{}, []string{"main", commandName}, os.Stdout)

	if err == nil {
		t.Fatalf("Expected error, got: nil")
	}

}
