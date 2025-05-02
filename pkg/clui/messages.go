package clui

import (
	"fmt"
	"io"
	"os"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"github.com/olekukonko/tablewriter"
)

// Normal returns a UIMessage that prints a normal message.
func (u *UI) Normal() *Message {
	return &Message{
		ui:           u,
		msgType:      normal,
		interactions: []interaction{},
		end:          -1,
	}
}

// Exclamation returns a UIMessage that prints an exclamation message.
func (u *UI) Exclamation() *Message {
	return &Message{
		ui:           u,
		msgType:      exclamation,
		interactions: []interaction{},
		end:          -1,
	}
}

// Note returns a UIMessage that prints a note message.
func (u *UI) Note() *Message {
	return &Message{
		ui:           u,
		msgType:      note,
		interactions: []interaction{},
		end:          -1,
	}
}

// Success returns a UIMessage that prints a success message.
func (u *UI) Success() *Message {
	return &Message{
		ui:           u,
		msgType:      success,
		interactions: []interaction{},
		end:          -1,
	}
}

// Problem returns a Message that prints a message that describes a problem.
func (u *UI) Problem() *Message {
	return &Message{
		ui:           u,
		msgType:      problem,
		interactions: []interaction{},
		end:          -1,
	}
}

// Msgf prints a formatted message on the CLI.
func (u *Message) Msgf(message string, a ...interface{}) {
	u.Msg(fmt.Sprintf(message, a...))
}

// Do is syntactic sugar for Msg("").
func (u *Message) Do() {
	u.Msg("")
}

// Msg prints a message on the CLI, resolving emoji as it goes.
func (u *Message) Msg(message string) {
	message = emoji.Sprint(message)

	// Print a newline before starting output, if not compact.
	if message != "" && !u.compact {
		fmt.Fprintln(u.ui.Output())
	}

	if !u.noNewline {
		message += "\n"
	}

	var output io.Writer

	switch u.msgType {
	case normal:
		output = u.ui.Output()
	case exclamation:
		output = u.ui.Output()
		message = color.YellowString(message)
	case note:
		output = u.ui.Output()
		message = color.BlueString(message)
	case success:
		output = u.ui.Output()
		message = color.GreenString(message)
	case progress:
		output = u.ui.Output()
	case problem:
		output = u.ui.Err()
		message = color.RedString(message)
	}

	fmt.Fprintf(output, "%s", message)

	for _, interaction := range u.interactions {
		switch interaction.variant {
		case ask:
			switch interaction.valueType { //nolint:exhaustive
			case tBool:
				*(interaction.value.(*bool)) = u.readBool(interaction.name, interaction.boolMap)
			case tInt:
				*(interaction.value.(*int64)) = u.readInt(interaction.name, interaction.allowedIntValues...)
			case tString:
				*(interaction.value.(*string)) = u.readString(interaction.name)
			case tPassword:
				*(interaction.value.(*string)) = u.readPassword(interaction.name, interaction.stdin)
			}
		case show:
			switch interaction.valueType { //nolint:exhaustive
			case tBool:
				fmt.Fprintf(output, "%s: %s\n", emoji.Sprint(interaction.name), color.MagentaString("%t", interaction.value))
			case tInt:
				fmt.Fprintf(output, "%s: %s\n", emoji.Sprint(interaction.name), color.CyanString("%d", interaction.value))
			case tString:
				fmt.Fprintf(output, "%s: %s\n", emoji.Sprint(interaction.name), color.GreenString("%s", interaction.value))
			case tErr:
				if u.stacks {
					fmt.Fprintf(output, "%s\n", color.RedString("%+v", interaction.value))
				} else {
					fmt.Fprintf(output, "%s\n", color.RedString("%v", interaction.value))
				}
			}
		}
	}

	for idx, headers := range u.table.headers {
		table := tablewriter.NewWriter(u.ui.output)
		table.SetHeader(headers)
		table.SetBorders(tablewriter.Border{Left: false, Top: false, Right: false, Bottom: false})
		table.SetBorder(false)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetHeaderLine(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetColumnSeparator("")
		table.SetAutoWrapText(!u.table.noAutoWrap)

		if idx < len(u.table.data) {
			table.AppendBulk(u.table.data[idx])
		}

		table.Render()
	}

	if u.end > -1 {
		os.Exit(u.end)
	}
}

// NoNewline disables the printing of a newline after a message output.
func (u *Message) NoNewline() *Message {
	u.noNewline = true

	return u
}

// Compact disables the printing of a newline before starting output.
func (u *Message) Compact() *Message {
	u.compact = true

	return u
}

// WithEnd ends the entire process after printing the message.
func (u *Message) WithEnd(code int) *Message {
	u.end = code
	return u
}

// WithStack causes error stacks to be printed.
func (u *Message) WithStacks() *Message {
	u.stacks = true
	return u
}

// WithBoolValue adds a bool value to be printed in the message.
func (u *Message) WithBoolValue(name string, value bool) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   show,
		valueType: tBool,
		value:     value,
	})

	return u
}

// WithStringValue adds a string value to be printed in the message.
func (u *Message) WithStringValue(name, value string) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   show,
		valueType: tString,
		value:     value,
	})

	return u
}

// WithErr adds an error value to be printed in the message.
func (u *Message) WithErr(err error) *Message {
	u.interactions = append(u.interactions, interaction{
		variant:   show,
		valueType: tErr,
		value:     err,
	})

	return u
}

// WithIntValue adds an int value to be printed in the message.
func (u *Message) WithIntValue(name string, value int64) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   show,
		valueType: tInt,
		value:     value,
	})

	return u
}
