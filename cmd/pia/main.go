package main

import (
	"fmt"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/crookdc/pia/squeak"
	"log"
	"strings"
)

func main() {
	p := tea.NewProgram(app())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func app() model {
	statement := textinput.New()
	statement.Placeholder = "let code = \"clean\";"
	statement.Focus()
	statement.Width = 100
	return model{
		evaluator: squeak.NewEvaluator(),
		statement: statement,
	}
}

type model struct {
	evaluator *squeak.Evaluator
	statement textinput.Model
	object    squeak.Object
	err       error
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			return m, repl(m.evaluator, m.statement.Value())
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		}
	case error:
		m.err = msg
		return m, nil
	case squeak.Object:
		m.err = nil
		m.object = msg
		return m, nil
	}
	m.statement, cmd = m.statement.Update(msg)
	return m, cmd
}

func (m model) View() string {
	sb := strings.Builder{}
	if m.err != nil {
		sb.WriteString(fmt.Sprintf("%+v\n", m.err))
	} else if m.object != nil {
		sb.WriteString(fmt.Sprintf("%+v\n", m.object))
	}
	sb.WriteString(m.statement.View())
	return sb.String()
}

func repl(ev *squeak.Evaluator, src string) tea.Cmd {
	return func() tea.Msg {
		lx, err := squeak.NewLexer(strings.NewReader(src))
		if err != nil {
			return err
		}
		plx, err := squeak.NewPeekingLexer(lx)
		if err != nil {
			return err
		}
		st, err := squeak.NewParser(plx).Next()
		if err != nil {
			return err
		}
		obj, err := ev.Statement(st)
		if err != nil {
			return err
		}
		return obj
	}
}
