package main

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
)

type model struct {
	choices  PackagesStr
	cursor   int
	selected map[int]struct{}
}

func (m model) len() int {
	var length int
	for _, v := range m.choices {
		length += len(v.Packages)
	}
	return length
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case ".":
			m.choices.install()
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < m.len()-1 {
				m.cursor++
			}
		case "enter", " ":
			_, ok := m.selected[m.cursor]
			if ok {
				delete(m.selected, m.cursor)
				m.choices.toggle(m.cursor, true)
			} else {
				m.selected[m.cursor] = struct{}{}
				m.choices.toggle(m.cursor, false)
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	s := "Please select the recipes to install:\n\n"
	var i int

	for _, v := range m.choices {
		for _, choice := range v.Packages {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			checked := "x"
			_, ok := m.selected[i]
			if ok {
				checked = " "
			}

			i++

			s1 := fmt.Sprintf("%s [%s] %s\n", cursor, checked, choice.User+"/"+choice.Repo)
			if ok {
				s1 = termenv.String(s1).Foreground(termenv.ColorProfile().Color("212")).String()
			}
			s += s1
		}
	}

	s += "\nPress . to install, q/ctrl+c to quit.\n"

	return s
}
