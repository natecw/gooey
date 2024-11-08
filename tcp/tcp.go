package tcp

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"sync"
	"syscall"
)

const (
	VERSION     = 1
	HEADER_SIZE = 4
)

type Server struct {
	conns       []Connection
	listener    net.Listener
	mutex       sync.RWMutex
	FromClients chan CmdWrapper
}

type Command struct {
	Command byte
	Data    []byte
}

type CmdWrapper struct {
	Cmd  *Command
	Conn *Connection
}

func (c *Command) MarshalBinary() ([]byte, error) {
	n := uint16(len(c.Data))
	lengthData := make([]byte, 2)
	binary.BigEndian.PutUint16(lengthData, n)

	b := make([]byte, 0, 4+n)
	b = append(b, VERSION)
	b = append(b, c.Command)
	b = append(b, lengthData...)
	return append(b, c.Data...), nil
}

func (c *Command) UnmarshalBinary(data []byte) error {
	if data[0] != VERSION {
		return fmt.Errorf("incorrect version: %d != %d", data[0], VERSION)
	}

	length := int(binary.BigEndian.Uint16(data[2:]))
	end := HEADER_SIZE + length

	if len(data) < end {
		return fmt.Errorf("not enough data to parse packet: got %d expected %d", len(data), HEADER_SIZE+length)
	}

	c.Command = data[1]
	c.Data = data[HEADER_SIZE:end]
	return nil
}

func (s *Server) Size() int {
	return len(s.conns)
}

func (s *Server) Close() {
	s.listener.Close()
}

func NewServer(port int16) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	return &Server{
		FromClients: make(chan CmdWrapper, 10),
		conns:       make([]Connection, 0, 10),
		listener:    ln,
		mutex:       sync.RWMutex{},
	}, nil
}

func (s *Server) Send(command *Command) {
	s.mutex.RLock()
	removals := make([]int, 0)
	slog.Debug("sending command", "msg", command)
	for i, conn := range s.conns {
		err := conn.Writer.Write(command)
		if err != nil {
			if errors.Is(err, syscall.EPIPE) {
				slog.Debug("connection closed by client", "id", i)
			} else {
				slog.Debug("connection error so removing", "id", i, "error", err)
			}
			removals = append(removals, i)
		}
	}
	s.mutex.RUnlock()

	if len(removals) > 0 {
		s.mutex.Lock()
		for i := len(removals) - 1; i >= 0; i-- {
			idx := removals[i]
			s.conns = append(s.conns[:idx], s.conns[idx+1:]...)
		}
		s.mutex.Unlock()
	}
}

func (s *Server) Start() {
	id := 0
	for {
		c, err := s.listener.Accept()
		if err != nil {
			slog.Error("error accepting connections", "error", err)
		}

		con := NewConnection(c, id)
		slog.Debug("new connection", "conn", con.Id)

		s.mutex.RLock()
		s.conns = append(s.conns, con)
		s.mutex.Unlock()

		id++
		go handleConnection(s, &con)
	}
}

func handleConnection(server *Server, conn *Connection) {
	for {
		cmd, err := conn.Next()
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Debug("connection EOF", "id", conn.Id, "error", err)
			} else {
				slog.Error("error reading from socket", "id", conn.Id, "error", err)
			}
		}

		server.FromClients <- CmdWrapper{Cmd: cmd, Conn: conn}
	}
}
