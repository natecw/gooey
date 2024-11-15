package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	con, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("error connecting to server", "error=", err)
		return
	}
	defer con.Close()

	go handleMessages(con)
	sender(con)
}

func handleMessages(con net.Conn) {
	scanner := bufio.NewScanner(con)
	for scanner.Scan() {
		msg := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Println("error reading from server", err)
			return
		}
		fmt.Println(msg)
	}
}

func sender(con net.Conn) {
	r := bufio.NewReader(os.Stdin)
	fmt.Printf("username: ")
	username, err := r.ReadString('\n')
	if username == "" {
		fmt.Println("username cannot be empty.")
		return
	}

	if err != nil {
		fmt.Println("error reading:", err)
		return
	}

	_, err = fmt.Fprintln(con, username)
	if err != nil {
		fmt.Println("error sending username", "username", username, "error", err)
		return
	}

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		msg := scanner.Text()
		if err := scanner.Err(); err != nil {
			fmt.Println("error reading from server", err)
			return
		}
		msg = strings.TrimSpace(msg)

		_, err = fmt.Fprintf(con, "%s: %s\n", username, msg)
		if err != nil {
			fmt.Println("error sending message", "error=", err)
		}
	}
}
