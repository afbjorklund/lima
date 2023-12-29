package uiutil

import (
	"errors"

	"github.com/charmbracelet/bubbles/list"
	 tea "github.com/charmbracelet/bubbletea"
	 "github.com/charmbracelet/lipgloss"
)

var InterruptErr = errors.New("interrupt")

var docStyle = lipgloss.NewStyle().Margin(1, 2)

type okitem struct {
	value bool
	title string
}

func (i okitem) Title() string       { return i.title }
func (i okitem) Description() string { return "" }
func (i okitem) FilterValue() string { return i.title }

type okmodel struct {
        list     list.Model
        choice   bool
        quitting bool
}

func (m okmodel) Init() tea.Cmd {
	return nil
}

func (m okmodel) View() string {
	return docStyle.Render(m.list.View())
}

func (m okmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(okitem)
			if ok {
				m.choice = i.value
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// Confirm is a regular text input that accept yes/no answers.
func Confirm(message string, defaultParam bool) (bool, error) {
	items := []list.Item{}
	items = append(items, okitem{value: true, title: "Yes"})
	items = append(items, okitem{value: false, title: "No"})

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = message
	l.SetShowStatusBar(false)

	m := okmodel{list: l}
	p := tea.NewProgram(m)
	r, err := p.Run()
	if err != nil {
		return false, err
        }
	if r.(okmodel).quitting {
		return false, InterruptErr
	}
	return r.(okmodel).choice, nil
}


type listitem struct {
	value int
	title string
}

func (i listitem) Title() string       { return i.title }
func (i listitem) Description() string { return "" }
func (i listitem) FilterValue() string { return i.title }

type listmodel struct {
        list     list.Model
        choice   int
        quitting bool
}

func (m listmodel) Init() tea.Cmd {
	return nil
}

func (m listmodel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch keypress := msg.String(); keypress {
		case "q", "ctrl+c":
			m.quitting = true
			return m, tea.Quit

		case "enter":
			i, ok := m.list.SelectedItem().(listitem)
			if ok {
				m.choice = i.value
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m listmodel) View() string {
	return docStyle.Render(m.list.View())
}

// Select is a prompt that presents a list of various options
// to the user for them to select using the arrow keys and enter.
func Select(message string, options []string) (int, error) {
	items := []list.Item{}
	for i, option := range options {
		items = append(items, listitem{value: i, title: option})
	}

	l := list.New(items, list.NewDefaultDelegate(), 0, 0)
	l.Title = message
	l.SetShowStatusBar(false)

	m := listmodel{list: l}
	p := tea.NewProgram(m)
	r, err := p.Run()
	if err != nil {
		return -1, err
        }
	if r.(listmodel).quitting {
		return -1, InterruptErr
	}
	return r.(listmodel).choice, nil
}
