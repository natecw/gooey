package main

import (
	"log"

	"github.com/natecw/gooey/tcp"
)

func main() {
	tcpServ, err := tcp.NewServer(8080)
	if err != nil {
		log.Fatalln("error starting server", "error=", err)
	}
	defer tcpServ.Close()

	tcpServ.Start()
}
