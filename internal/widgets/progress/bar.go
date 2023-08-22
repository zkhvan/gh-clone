package progress

import (
	"errors"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
)

var ErrAborted = errors.New("program aborted")

type Bar struct {
	program *tea.Program
}

func New() *Bar {
	m := model{
		activeMsgs: make(map[string]struct{}),
		progress:   progress.New(),
	}

	program := tea.NewProgram(m)

	return &Bar{program: program}
}

func (b *Bar) Wait() error {
	m, err := b.program.Run()
	if m, ok := m.(model); ok {
		if m.abort {
			return ErrAborted
		}
	}
	return err
}

func (b *Bar) SetTotal(total int) {
	b.program.Send(totalUpdateMsg(total))
}

func (b *Bar) Start(msg string) {
	b.program.Send(startMsg(msg))
}

func (b *Bar) Increment(msg string) {
	b.program.Send(progressMsg(msg))
}
