package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/adirg/gosmosis/server"
)

type Client struct {
	host    string
	port    uint
	workDir string
	wg      sync.WaitGroup
}

func NewClient(host string, port uint) *Client {
	c := new(Client)

	c.host = host
	c.port = port

	return c
}

func (c *Client) setWorkDir(workDir string) {
	absWorkDir, err := filepath.Abs(workDir)
	if err != nil {
		log.Fatal("Error determining absolute path of directory: ", err.Error())
	}

	mode, err := os.Stat(absWorkDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	if !mode.IsDir() {
		log.Fatalf("%s is not a directory", absWorkDir)
	}

	c.workDir = absWorkDir
}

func (c *Client) connect() net.Conn {
	connectionString := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		log.Fatal(err.Error())
	}

	return conn
}

func set(conn net.Conn, r io.Reader, size int64, hash []byte) {
	conn.Write([]byte{server.OpSet}) // opcode
	conn.Write(hash)                 // hash

	sizeBuf := make([]byte, 8)
	binary.PutVarint(sizeBuf, size)
	binary.Write(conn, binary.LittleEndian, size)

	buf := make([]byte, 1024)
	io.CopyBuffer(conn, r, buf)
}

func get(conn net.Conn, w io.Writer, hash []byte) {
	conn.Write([]byte{server.OpGet}) // opcode
	conn.Write(hash)

	var size int64
	binary.Read(conn, binary.LittleEndian, &size)
	log.Println("Size of manifest file: ", size)

	r := io.LimitReader(conn, size)
	buf := make([]byte, 1024)
	_, err := io.CopyBuffer(w, r, buf)
	if err != nil {
		log.Println("Failed to read manifest content")
	}

}

func setLabel(conn net.Conn, label string, hash []byte) {
	conn.Write([]byte{server.OpSetLabel}) // opcode
	conn.Write(hash[:])                   // hash

	labelBuf := []byte(label)
	size := int64(len(labelBuf))
	sizeBuf := make([]byte, 8)
	binary.PutVarint(sizeBuf, size)
	binary.Write(conn, binary.LittleEndian, size)
	conn.Write(labelBuf)
}

func getLabel(conn net.Conn, label string) []byte {
	conn.Write([]byte{server.OpGetLabel}) // opcode

	// send label
	labelBuf := []byte(label)
	size := int64(len(labelBuf))
	binary.Write(conn, binary.LittleEndian, size)
	conn.Write(labelBuf)

	// get manifest hash
	hash := make([]byte, 32)
	_, err := io.ReadFull(conn, hash)
	if err != nil {
		log.Println("Failed to read hash of label: ", label)
		return nil
	}

	return hash
}
