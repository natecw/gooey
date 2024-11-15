package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
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
		register:   make(chan Client, 1),
		unregister: make(chan net.Conn, 1),
		broadcast:  make(chan string, 1),
	}
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.conn] = client
			s.mu.Unlock()
			msg := fmt.Sprintf("%s has joined the chat\n", client.username)
			fmt.Println(msg)
			s.broadcast <- msg
		case con := <-s.unregister:
			s.mu.Lock()
			client, ok := s.clients[con]
			if ok {
				delete(s.clients, con)
				msg := fmt.Sprintf("%s has left the chat\n", client.username)
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

	lsnr, err := net.Listen("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error starting listener", "error", err)
		os.Exit(1)
	}
	defer lsnr.Close()

	fmt.Println("Server started.")
	for {
		con, err := lsnr.Accept()
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

	reader := bufio.NewReader(con)
	name, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("error reading username: %v\n", err)
		return
	}

	name = strings.TrimSpace(name)
	if name == "" {
		name = "Anonymous"
	}

	client := Client{conn: con, username: name}
	server.register <- client

	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("error reading connection: %v\n", err)
			return
		}
		if msg != "" {
			message := strings.TrimSpace(msg)
			fmt.Println(message)
			server.broadcast <- message
		}
	}
}
