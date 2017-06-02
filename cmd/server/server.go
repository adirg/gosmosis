package main

import "flag"
import "github.com/adirg/gosmosis/server"

func main() {
	var port = flag.Uint("port", 8080, "the port to listen on")

	flag.Parse()

	server := &server.Server{}
	server.Listen(*port)
}
