package clui

import (
	"io"
	"os"
	"runtime"

	"github.com/fatih/color"
	"github.com/opcr-io/policy/pkg/x"
)

type (
	msgType      int
	valueType    int
	valueVariant int
)

const (
	normal msgType = iota
	exclamation
	problem
	note
	success
	progress
)

const (
	tBool valueType = iota
	tString
	tInt
	tErr
	tPassword
)

const (
	show valueVariant = iota
	ask
)

// UI contains functionality for dealing with the user on the CLI.
type UI struct {
	output io.Writer
	err    io.Writer
	input  io.Reader
}

// Message represents a piece of information we want displayed to the user.
type Message struct {
	ui           *UI // For access to requested verbosity.
	msgType      msgType
	end          int
	compact      bool
	noNewline    bool
	stacks       bool
	interactions []interaction
	table        table
}

type table struct {
	headers    [][]string
	data       [][][]string
	noAutoWrap bool
}

type interaction struct {
	variant          valueVariant
	valueType        valueType
	name             string
	value            interface{}
	boolMap          map[string]bool
	stdin            bool
	allowedIntValues []int
}

// NewUI creates a new UI.
func NewUI() *UI {
	if runtime.GOOS == x.IsWindows {
		return NewUIWithOutput(color.Output)
	} else {
		return NewUIWithOutput(os.Stdout)
	}
}

// NewUI creates a new UI with a specific output.
func NewUIWithOutput(output io.Writer) *UI {
	return NewUIWithOutputAndInput(output, os.Stdin)
}

func NewUIWithOutputAndInput(output io.Writer, input io.Reader) *UI {
	if runtime.GOOS == x.IsWindows {
		return NewUIWithOutputErrorAndInput(output, color.Error, input)
	}

	return NewUIWithOutputErrorAndInput(output, os.Stderr, input)
}

// NewUI creates a new UI with a specific input, error output and output.
func NewUIWithOutputErrorAndInput(output, err io.Writer, input io.Reader) *UI {
	return &UI{
		output: output,
		err:    err,
		input:  input,
	}
}

// Input returns the io.Reader used to read user input.
func (u *UI) Input() io.Reader {
	return u.input
}

// Output returns the io.Write used to print.
func (u *UI) Output() io.Writer {
	return u.output
}

func (u *UI) Err() io.Writer {
	return u.err
}
