package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"dio.wtf/joycontrol/joycontrol"
	tea "github.com/charmbracelet/bubbletea"
	"golang.org/x/exp/slices"
)

type model struct {
	action     []string
	current    string
	lastAction string

	controller *joycontrol.Controller
}

func initialModel(controller *joycontrol.Controller) model {
	return model{
		action:     []string{"A", "B", "X", "Y", "L", "ZL", "R", "ZR", "HOME", "UP", "DOWN", "LEFT", "RIGHT"},
		current:    "",
		lastAction: "",

		controller: controller,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Send() {
	m.controller.Press(m.current)
	time.Sleep(time.Second / 10)
	m.controller.Release(m.current)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		key := strings.ToUpper(msg.String())
		switch {
		case slices.Contains(m.action, key):
			m.lastAction = m.current
			m.current = key
			return m, nil

		case key == "CTRL+C", key == "Q":
			return m, tea.Quit

		case key == "ENTER":
			go m.Send()
		}
	}

	return m, nil
}

func (m model) View() string {
	// The header
	// s := "Type next action?\n\n"

	// s += fmt.Sprintf("last action: %s\n", m.lastAction)
	// s += strings.ToUpper(m.current)

	// s += "\n\nPress q to quit.\n"

	return ""
}

func main() {
	controller := joycontrol.NewController()
	server := joycontrol.NewServer(controller)
	server.Start()
	defer server.Stop()

	p := tea.NewProgram(initialModel(controller))
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
