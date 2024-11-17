package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Client struct {
	conn     net.Conn
	username string
}

type Server struct {
	clients    map[net.Conn]Client
	register   chan Client
	unregister chan net.Conn
	broadcast  chan string
	mu         sync.Mutex
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[net.Conn]Client),
		register:   make(chan Client),
		unregister: make(chan net.Conn),
		broadcast:  make(chan string, 5),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.conn] = client
			s.mu.Unlock()
			msg := fmt.Sprintf("%s has joined the chat", client.username)
			fmt.Println(msg)
			s.broadcast <- msg
		case con := <-s.unregister:
			s.mu.Lock()
			client, ok := s.clients[con]
			if ok {
				delete(s.clients, con)
				msg := fmt.Sprintf("%s has left the chat", client.username)
				fmt.Println(msg)
				s.broadcast <- msg
			}
			s.mu.Unlock()
			con.Close()
		case msg := <-s.broadcast:
			s.mu.Lock()
			for _, client := range s.clients {
				_, err := fmt.Fprintln(client.conn, msg)
				if err != nil {
					fmt.Println("error sending message", "user", client.username, "error", err.Error())
				}
			}
			s.mu.Unlock()
		}
	}
}

func main() {
	server := NewServer()
	go server.Run()

	listener, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error starting listener", "error", err)
		return
	}
	defer listener.Close()

	fmt.Println("Server started.")
	for {
		con, err := listener.Accept()
		if err != nil {
			fmt.Println("error accepting", "error", err)
			continue
		}
		fmt.Println("new connection", "url", con.RemoteAddr().String())
		go handleConnection(server, con)
	}
}

func handleConnection(server *Server, con net.Conn) {
	defer func() {
		server.unregister <- con
	}()

	client := Client{conn: con, username: "Anonymous"}
	server.register <- client

	reader := bufio.NewReader(con)
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading connection: %v\n", err)
			break
		}
		message := strings.TrimSpace(msg)
		if message == "" {
			continue
		}

		var cmd command
		parts := strings.Split(message, " ")
		switch parts[0] {
		case "/username":
			client.username = parts[1]
			cmd = command{
				cmd:       "/username",
				body:      parts[1],
				username:  client.username,
				broadcast: false,
			}
		default:
			cmd = command{
				cmd:       "/m",
				body:      message,
				username:  client.username,
				broadcast: true,
			}
		}

		if cmd.broadcast {
			fmt.Println(cmd.String())
			server.broadcast <- cmd.String()
		}
	}
}

type command struct {
	cmd       string
	body      string
	username  string
	broadcast bool
}

func (c *command) String() string {
	switch c.cmd {
	case "/username":
		return fmt.Sprintf("%s is now %s\n", c.username, c.body)
	case "/m":
		return fmt.Sprintf("%s: %s", c.username, c.body)
	default:
		return ""
	}
}
