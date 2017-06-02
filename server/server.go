package server

import "encoding/binary"
import "fmt"
import "log"
import "net"

type Server struct {
}

func (server *Server) Listen(port uint) {
	var bind_address = fmt.Sprintf(":%d", port)
	listener, err := net.Listen("tcp", bind_address)
	if err != nil {
		log.Fatal(err)
	}

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal("Error while accepting connection")
		}
		defer conn.Close()

		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	fmt.Println("Got connection: ", conn.RemoteAddr())

	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Closing the connection: ", conn.RemoteAddr())
		return
	}
	fmt.Printf("Got message: %x\n", buf[:n])

	size, _ := binary.Uvarint(buf[0:2])
	fmt.Println("Size: ", size)
	filename := string(buf[2 : size+2])

	fmt.Println("Filename: ", filename)
	fmt.Printf("Hash: %x\n", buf[size+2:n])
}
