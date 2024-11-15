package main

import (
	"fmt"
	"log"
	"net"

	tea "github.com/charmbracelet/bubbletea"
)

type model struct {
	input string
}

func main() {
	con, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error connecting to server", "error=", err)
		return
	}
	defer con.Close()

	p := tea.NewProgram(initiate())
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initiate() model {
	return model{input: "teehee"}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return nil, nil
}

func (m model) View() string {
	return ""
}
