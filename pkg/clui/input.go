package clui

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
	"syscall"

	"github.com/fatih/color"
	"github.com/kyokomi/emoji"
	"golang.org/x/term"
)

// WithAskBool waits for the user's input for a boolean value.
func (u *Message) WithAskBool(name string, result *bool) *Message {
	return u.WithAskBoolMap(name, result, map[string]bool{
		"true":  true,
		"false": false,
	})
}

// WithAskBool waits for the user's input for a boolean value.
func (u *Message) WithAskBoolMap(name string, result *bool, answerMap map[string]bool) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   ask,
		valueType: tBool,
		value:     result,
		boolMap:   answerMap,
	})

	return u
}

// WithAskString waits for the user's input for a string value.
func (u *Message) WithAskString(name string, result *string) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   ask,
		valueType: tString,
		value:     result,
	})

	return u
}

// WithAskInt waits for the user's input for an int value.
func (u *Message) WithAskInt(name string, result *int64, allowedValues ...int) *Message {
	u.interactions = append(u.interactions, interaction{
		name:             name,
		variant:          ask,
		valueType:        tInt,
		value:            result,
		allowedIntValues: allowedValues,
	})

	return u
}

// WithAskPassword waits for the user's input for a password value.
func (u *Message) WithAskPassword(name string, stdin bool, result *string) *Message {
	u.interactions = append(u.interactions, interaction{
		name:      name,
		variant:   ask,
		valueType: tPassword,
		value:     result,
		stdin:     stdin,
	})

	return u
}

func (u *Message) readBool(message string, boolMap map[string]bool) bool {
	if !strings.HasSuffix(message, "?") && !strings.HasSuffix(message, ":") {
		message += ":"
	}

	if len(boolMap) == 0 {
		boolMap = map[string]bool{
			"true":  true,
			"false": false,
		}
	}

	scanner := bufio.NewScanner(u.ui.input)

	for {
		if u.compact {
			fmt.Fprintf(u.ui.Output(), "> %s ", emoji.Sprint(message))
		} else {
			fmt.Fprintf(u.ui.Output(), "> [%s] %s ", color.MagentaString("bool"), emoji.Sprint(message))
		}

		scanner.Scan()
		text := scanner.Text()

		value, ok := boolMap[strings.ToLower(text)]

		if !ok {
			u.ui.Problem().WithStringValue("  input", text).Msg("Invalid value.")
			continue
		}

		return value
	}
}

func (u *Message) readString(message string) string {
	if !strings.HasSuffix(message, "?") && !strings.HasSuffix(message, ":") {
		message += ":"
	}

	if u.compact {
		fmt.Fprintf(u.ui.Output(), "> %s ", emoji.Sprint(message))
	} else {
		fmt.Fprintf(u.ui.Output(), "> [%s] %s ", color.GreenString("text"), emoji.Sprint(message))
	}

	scanner := bufio.NewScanner(u.ui.input)
	scanner.Scan()
	value := scanner.Text()

	return value
}

func (u *Message) readInt(message string, allowedValues ...int) int64 {
	if !strings.HasSuffix(message, "?") && !strings.HasSuffix(message, ":") {
		message += ":"
	}

	var (
		result int64
		err    error
	)

	scanner := bufio.NewScanner(u.ui.input)

	for {
		if u.compact {
			fmt.Fprintf(u.ui.Output(), "> %s ", emoji.Sprint(message))
		} else {
			fmt.Fprintf(u.ui.Output(), "> [%s] %s ", color.CyanString("integer"), emoji.Sprint(message))
		}

		scanner.Scan()
		text := scanner.Text()

		result, err = strconv.ParseInt(text, 10, 64)
		if err != nil {
			u.ui.Problem().WithStringValue("  input", text).Msg("Value is not an integer.")
			continue
		}

		if len(allowedValues) > 0 {
			found := false

			for _, value := range allowedValues {
				if int64(value) == result {
					found = true
					break
				}
			}

			if !found {
				u.ui.Problem().WithStringValue("  input", text).Msg("Value is not an allowed option.")
				continue
			}
		}

		return result
	}
}

func (u *Message) readPassword(message string, stdin bool) string {
	if !strings.HasSuffix(message, "?") && !strings.HasSuffix(message, ":") {
		message += ":"
	}

	var value string

	if stdin {
		contents, err := io.ReadAll(u.ui.input)
		if err != nil {
			u.ui.Problem().WithStringValue("  input", err.Error()).Msg("failed to read password from stdin")
			return ""
		}

		value = strings.TrimSuffix(string(contents), "\n")
		value = strings.TrimSuffix(value, "\r")
	}

	if value == "" {
		fmt.Fprintf(u.ui.Output(), "> [%s] %s ", color.GreenString("password"), emoji.Sprint(message))

		byteValue, err := term.ReadPassword(syscall.Stdin)
		if err != nil {
			u.ui.Problem().WithStringValue("  input", err.Error()).Msg("failed to read password")
			return ""
		}

		// assume that go knows how to handle newline on nix platforms: https://golang.org/src/fmt/print.go
		if _, err := u.ui.output.Write([]byte{'\n'}); err != nil {
			u.ui.Problem().WithStringValue("  input", err.Error()).Msg("failed to write to output")
			return ""
		}

		value = string(byteValue)
	}

	return value
}
