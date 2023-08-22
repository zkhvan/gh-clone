package progress

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/progress"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#626262")).Render

type startMsg string
type progressMsg string

type totalUpdateMsg int

func finalPause() tea.Cmd {
	return tea.Tick(time.Millisecond*750, func(_ time.Time) tea.Msg {
		return nil
	})
}

type model struct {
	activeMsgs map[string]struct{}

	current int
	total   int

	progress progress.Model

	abort bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		m.abort = true
		return m, tea.Quit

	case tea.WindowSizeMsg:
		m.progress.Width = msg.Width - 2*2 - 4
		if m.progress.Width > 80 {
			m.progress.Width = 80
		}
		return m, nil

	case totalUpdateMsg:
		m.total = int(msg)
		return m, nil

	case startMsg:
		m.activeMsgs[string(msg)] = struct{}{}
		return m, nil

	case progressMsg:
		var cmds []tea.Cmd

		if len(string(msg)) > 0 {
			delete(m.activeMsgs, string(msg))
		}

		m.current++
		if m.current >= m.total {
			cmds = append(cmds, tea.Sequence(finalPause(), tea.Quit))
		}

		cmds = append(cmds, m.progress.SetPercent(float64(m.current)/float64(m.total)))
		return m, tea.Batch(cmds...)

	case progress.FrameMsg:
		progressModel, cmd := m.progress.Update(msg)
		m.progress = progressModel.(progress.Model)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	pad := strings.Repeat(" ", 2)

	var msgs []string
	for msg := range m.activeMsgs {
		msgs = append(msgs, msg)
	}

	progress := fmt.Sprintf("[%d/%d]", m.current, m.total)

	return "\n" +
		pad + progress + " " + strings.Join(msgs, ", ") + "\n" +
		pad + m.progress.View() + "\n\n" +
		pad + helpStyle("Press any key to quit")
}
