package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/natecw/gooey/common"
)

func main() {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:8080")
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}

	con, err := net.DialTCP("tcp", nil, addr)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer con.Close()
	go printMessages(con)
	send(con)
}

func printMessages(con *net.TCPConn) {
	for {
		msg, err := common.Read(con)
		if err == io.EOF {
			con.Close()
			fmt.Println("Connection closed")
			os.Exit(0)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(msg)
	}
}

func send(con *net.TCPConn) {
	fmt.Print("Enter username: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	username = username[:len(username)-1]
	fmt.Println(username)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("Enter text: ")
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}
		err = common.Write(con, username+": "+msg)
		if err != nil {
			log.Println(err)
		}
	}
}
