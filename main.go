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
	if checkinArgs.NArg() < 2 {
		log.Fatal("Missing directory / label")
	}

	dir := checkinArgs.Arg(0)
	label := checkinArgs.Arg(1)
	client.Checkin(*host, *port, dir, label)
}

func handleCheckoutCmd(args []string) {
	checkoutArgs := flag.NewFlagSet("checkout", flag.ExitOnError)
	host := checkoutArgs.String("host", "127.0.0.1", "connect to ip address")
	port := checkoutArgs.Uint("port", 3333, "connect to port")

	checkoutArgs.Parse(args)
	if checkoutArgs.NArg() < 2 {
		log.Fatal("Missing directory / label")
	}

	dir := checkoutArgs.Arg(0)
	label := checkoutArgs.Arg(1)
	client.Checkout(*host, *port, dir, label)
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
		handleCheckoutCmd(os.Args[2:])
	case "list-labels":
	case "rm-label":
	default:
		// TODO: Print usage
		log.Fatal("Invalid command: ", os.Args[1])
	}
}
