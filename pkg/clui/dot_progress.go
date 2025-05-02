package clui

import (
	"fmt"
	"sync"
	"time"

	"github.com/fatih/color"
)

// DotProgress implements a progress indicator
// by printing a series of dots as time goes by.
type DotProgress struct {
	ui       *UI
	mu       *sync.RWMutex //
	Delay    time.Duration // Delay is the speed of the indicator
	active   bool          // active holds the state of the spinner
	stopChan chan struct{} // stopChan is a channel used to stop the indicator.
}

// Standard values for the current dot-based progress.
const dotTime = 1 * time.Second

// NewDotProgress creates a new DotProgress.
func NewDotProgress(ui *UI, message string) *DotProgress {
	message = finalMsg(message)
	p := &DotProgress{
		ui:       ui,
		Delay:    dotTime,
		mu:       &sync.RWMutex{},
		active:   false,
		stopChan: make(chan struct{}, 1),
	}

	p.ui.ProgressNote().NoNewline().Msg(message)

	p.Start()

	return p
}

// Start start.
func (p *DotProgress) Start() {
	p.mu.Lock()
	if p.active {
		p.mu.Unlock()
		return
	}

	p.active = true
	p.mu.Unlock()

	go func() {
		for {
			select {
			case <-p.stopChan:
				return
			default:
				p.mu.Lock()
				if !p.active {
					p.mu.Unlock()
					return
				}

				p.ui.Normal().
					Compact().
					NoNewline().
					Msg(color.MagentaString("."))

				delay := p.Delay

				p.mu.Unlock()

				time.Sleep(delay)
			}
		}
	}()
}

// Stop stops displaying the dot progress spinner.
func (p *DotProgress) Stop() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.active {
		p.active = false
		p.ui.Normal().Compact().Msg("")
		p.stopChan <- struct{}{}
	}
}

// ChangeMessagef extends the dot-based progress with the ability to
// change the message mid-flight.
func (p *DotProgress) ChangeMessagef(message string, a ...interface{}) {
	p.ChangeMessage(fmt.Sprintf(message, a...))
}

// ChangeMessage extends the dot-based progress with the ability to
// change the message mid-flight.
func (p *DotProgress) ChangeMessage(message string) {
	message = finalMsg(message)

	p.Stop()

	p.ui.ProgressNote().NoNewline().Msg(message)

	p.Start()
}

func finalMsg(message string) string {
	return message + " "
}
