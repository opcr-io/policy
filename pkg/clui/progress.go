package clui

import "fmt"

// Progress abstracts the operations for a progress meter and/or
// spinner used to indicate the cli waiting for some background
// operation to complete.
type Progress interface {
	Start()
	Stop()
	ChangeMessage(message string)
	ChangeMessagef(message string, a ...interface{})
}

// Progressf creates, configures, and returns an active progress
// meter. It accepts a formatted message.
func (u *UI) Progressf(message string, a ...interface{}) Progress {
	return u.Progress(fmt.Sprintf(message, a...))
}

// Progress creates, configures, and returns an active progress
// meter. It accepts a fixed message.
func (u *UI) Progress(message string) Progress {
	return NewDotProgress(u, message)
}

// ProgressNote returns a UIMessage that prints a progress-related message.
func (u *UI) ProgressNote() *Message {
	return &Message{
		ui:           u,
		msgType:      progress,
		interactions: []interaction{},
		end:          -1,
	}
}
