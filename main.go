package main

import (
	"flag"
	"log"
	"os"

	"github.com/adirg/gosmosis/client"
	"github.com/adirg/gosmosis/server"
)

func handleServerCmd(args []string) {
	serverArgs := flag.NewFlagSet("server", flag.ExitOnError)
	host := serverArgs.String("host", "0.0.0.0", "bind to ip address")
	port := serverArgs.Uint("port", 3333, "listen on port")
	path := serverArgs.String("path", "/var/lib/osmosis", "local path data store")

	serverArgs.Parse(args)
	s := server.NewServer(*host, *port, *path)
	s.Start()
}

func handleCheckinCmd(args []string) {
	checkinArgs := flag.NewFlagSet("checkin", flag.ExitOnError)
	host := checkinArgs.String("host", "127.0.0.1", "connect to ip address")
	port := checkinArgs.Uint("port", 3333, "connect to port")

	checkinArgs.Parse(args)
	if checkinArgs.NArg() == 0 {
		log.Fatal("Missing directory to checkin")
	}

	dir := checkinArgs.Arg(0)
	client.Checkin(*host, *port, dir)
}

func main() {
	if len(os.Args) < 2 {
		// TODO: Print usage
		log.Fatal("Missing command")
	}

	switch os.Args[1] {
	case "server":
		handleServerCmd(os.Args[2:])
	case "checkin":
		handleCheckinCmd(os.Args[2:])
	case "checkout":
	case "list-labels":
	case "rm-label":
	default:
		// TODO: Print usage
		log.Fatal("Invalid command: ", os.Args[1])
	}
}
