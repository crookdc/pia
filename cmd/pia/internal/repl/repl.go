package repl

import (
	"bytes"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crookdc/pia/squeak"
	"strings"
)

func Run() error {
	_, err := tea.NewProgram(initial(), tea.WithAltScreen()).Run()
	return err
}

func initial() model {
	prompt := textinput.New()
	prompt.Focus()
	out := bytes.NewBufferString("")
	return model{
		out:    out,
		in:     squeak.NewInterpreter(out),
		prompt: prompt,
	}
}

type model struct {
	console string
	out     *bytes.Buffer
	in      *squeak.Interpreter
	prompt  textinput.Model
	err     error
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			n, err := squeak.ParseString(m.prompt.Value())
			if err != nil {
				m.err = err
				break
			}
			err = m.in.Execute(n)
			if err != nil {
				m.err = err
			}
			m.prompt.SetValue("")
		}
	}
	out := m.out.String()
	if out != "" {
		m.console += out + "\n"
		m.out.Reset()
	}
	if m.err != nil {
		m.console += "[!] " + m.err.Error() + "\n"
		m.err = nil
	}
	m.prompt, cmd = m.prompt.Update(msg)
	return m, cmd
}

func (m model) View() string {
	sb := strings.Builder{}
	sb.WriteString(m.console)
	sb.WriteString(m.prompt.View())
	return sb.String()
}
