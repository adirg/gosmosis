package main

import "flag"
import "fmt"
import "github.com/adirg/gosmosis/client"

func main() {
	var server_ip = flag.String("server-ip", "127.0.0.1", "the ip of the server")
	var server_port = flag.Uint("server-port", 8080, "the port of the server")

	flag.Parse()
	fmt.Println("Server IP: ", *server_ip)
	fmt.Println("Server Port: ", *server_port)

	filename := flag.Arg(0)
	client := &client.Client{}
	client.Connect(*server_ip, *server_port)
	client.Upload(filename)
	client.Disconnect()
}
