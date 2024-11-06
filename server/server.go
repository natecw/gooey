package main

import (
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"

	"github.com/natecw/gooey/common"
)

var connections []net.Conn

func main() {
	url := "localhost:8080"
	ln, err := net.Listen("tcp", url)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
	defer ln.Close()

	fmt.Println("Listening on", url)

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}
		connections = append(connections, conn)
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	for {
		msg, err := common.Read(conn)
		if err != nil {
			if err == io.EOF {
				removeConnection(conn)
				conn.Close()
				return
			}
			slog.Warn(err.Error())
			return
		}
		fmt.Println(msg)
		fmt.Printf("Message received: %s\n", msg)
		broadcast(conn, msg)
	}
}

func removeConnection(conn net.Conn) {
	var i int
	for i = range connections {
		if connections[i] == conn {
			break
		}
	}
	connections = append(connections[:i], connections[i+1:]...)
}

func broadcast(conn net.Conn, msg string) {
	for i := range connections {
		if connections[i] != conn {
			err := common.Write(connections[i], msg)
			if err != nil {
				slog.Error(err.Error())
			}
		}
	}
}
