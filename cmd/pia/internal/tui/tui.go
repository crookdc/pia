package tui

import (
	"github.com/charmbracelet/bubbles/table"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var style = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color("240"))

func Run(wd string) error {
	_, err := tea.NewProgram(model{
		wd: wd,
		transactions: table.New(
			table.WithColumns([]table.Column{
				{
					Title: "Transaction",
					Width: 50,
				},
			}),
			table.WithRows([]table.Row{
				{"users/read/tx.yml"},
				{"users/create/tx.yml"},
				{"users/update/tx.yml"},
				{"users/delete/tx.yml"},
			}),
		),
	}).Run()
	return err
}

type model struct {
	wd           string
	transactions table.Model
	history      table.Model
	prompt       textinput.Model
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.transactions.SetHeight(msg.Height - 4)
		m.transactions.SetWidth((msg.Width - 4) / 3)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyCtrlT:
			m.transactions.Focus()
		case tea.KeyEnter:
			return m, tea.Batch(
				tea.Printf("%s", m.transactions.SelectedRow()),
			)
		}
	}
	m.transactions, cmd = m.transactions.Update(msg)
	return m, cmd
}

func (m model) View() string {
	return style.Render(m.transactions.View()) + "\n"
}
