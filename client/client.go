package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"strings"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	viewport viewport.Model
	messages []string
	textarea textarea.Model
	style    lipgloss.Style
	err      error
	conn     net.Conn
	inbound  chan message
}

type message string

func (m message) String() string {
	return string(m)
}

func main() {
	con, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error connecting to server", "error=", err)
		return
	}
	defer con.Close()

	p := tea.NewProgram(initiate(con))
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func initiate(con net.Conn) model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	ta.Prompt = "| "
	ta.CharLimit = 280
	ta.SetWidth(30)
	ta.SetHeight(1)
	ta.MaxHeight = 3

	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()
	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
	Type a message and press Enter to send`)

	ta.KeyMap.InsertNewline.SetEnabled(false)
	return model{
		viewport: vp,
		textarea: ta,
		messages: []string{},
		style:    lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:      nil,
		conn:     con,
		inbound:  make(chan message),
	}
}

func (m model) Init() tea.Cmd {
	go m.readMessages(m.inbound, m.conn)
	return tea.Batch(
		textarea.Blink,
		waitForMessages(m.inbound),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "esc", ":ctrl+c":
			fmt.Println(m.textarea.Value())
			return m, tea.Quit
		case "enter":
			v := m.textarea.Value()
			if v == "" {
				return m, nil
			}

			fmt.Fprintln(m.conn, v)
			m.messages = append(m.messages, m.style.Render("You: ")+v)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			return m, nil
		default:
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}
	case message:
		m.messages = append(m.messages, fmt.Sprintln(msg.String()))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.textarea.Reset()
		m.viewport.GotoBottom()
		return m, waitForMessages(m.inbound)
	case cursor.BlinkMsg:
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	default:
		return m, nil
	}
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func waitForMessages(msgs chan message) tea.Cmd {
	return func() tea.Msg {
		return message(<-msgs)
	}
}

func (m *model) readMessages(msgs chan message, conn net.Conn) {
	r := bufio.NewReader(conn)
	for {
		msg, err := r.ReadString('\n')
		if err != nil {
			log.Fatal("error reading connection", err)
		}
		if msg != "" {
			msgs <- message(msg)
		}
	}
}
