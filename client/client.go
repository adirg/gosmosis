package client

import "crypto/sha1"
import "encoding/binary"
import "fmt"
import "io"
import "log"
import "net"
import "os"

type Client struct {
	conn net.Conn
}

func (client *Client) Connect(host string, port uint) {
	var connection_string = fmt.Sprintf("%s:%d", host, port)
	fmt.Println("Connection string: ", connection_string)

	conn, err := net.Dial("tcp", connection_string)
	if err != nil {
		log.Fatal(err)
	}

	client.conn = conn
}

func (client *Client) Disconnect() {
	client.conn.Close()
}

func (client *Client) Upload(filename string) {
	fmt.Println("Uploading file: ", filename)

	file, err := os.Open(filename)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	h := sha1.New()
	_, err = io.Copy(h, file)
	if err != nil {
		log.Fatal(err)
	}

	data := h.Sum(nil)
	size := make([]byte, 2)

	binary.PutUvarint(size, uint64(len(filename)))

	client.conn.Write(size)
	client.conn.Write([]byte(filename))
	client.conn.Write(data)
	fmt.Printf("% x\n", data)
}
