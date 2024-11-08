package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand/v2"
	"net"
	"os"

	"github.com/natecw/gooey/commands"
	"github.com/natecw/gooey/tcp"
)

type opts struct {
	host string
	port uint
	id   int
}

func parseOptions() opts {
	opts := opts{}
	flag.UintVar(&opts.port, "port", 8080, "port of the server")
	flag.IntVar(&opts.id, "id", 0, "users id")
	flag.StringVar(&opts.host, "host", "localhost", "address of the server")
	flag.Parse()
	return opts
}

func main() {
	options := parseOptions()
	con, err := net.Dial("tcp", fmt.Sprintf("%s:%d", options.host, options.port))
	if err != nil {
		slog.Warn("error connecting to server", "server=", err)
	}
	defer con.Close()

	id := rand.Int()
	newCon := tcp.NewConnection(con, id)

	go printMessages(&newCon)
	send(&newCon)
}

func printMessages(con *tcp.Connection) {
	for {
		data, err := con.Reader.Read()
		if err == io.EOF {
			con.Close()
			fmt.Println("Connection closed")
			os.Exit(0)
		}
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(data)
	}
}

func send(con *tcp.Connection) {
	fmt.Print("Enter username: ")
	reader := bufio.NewReader(os.Stdin)
	username, err := reader.ReadString('\n')
	username = username[:len(username)-1]

	if err != nil {
		log.Fatal(err)
	}
	fmt.Print("> ")
	for {
		msg, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		cmd := tcp.Command{
			Command: byte(commands.Message),
			Data:    []byte(fmt.Sprintf("%s: %s", username, msg)),
		}
		err = con.Writer.Write(&cmd)
		if err != nil {
			slog.Error("error sending to server", "error", err, "cmd", cmd)
		}
	}
}
