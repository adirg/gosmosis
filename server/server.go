package server

import (
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
)

const (
	OP_SET = iota
	OP_GET
	OP_EXISTS
	OP_SET_LABEL
)

type Server struct {
	host       string
	port       uint
	rootDir    string
	objectsDir string
	draftsDir  string
	labelsDir  string
}

func NewServer(host string, port uint, rootDir string) *Server {
	s := new(Server)

	s.host = host
	s.port = port
	s.createDirs(rootDir)

	return s
}

func (s *Server) Start() {
	// Listen for incomming connections
	connectionString := fmt.Sprintf("%s:%d", s.host, s.port)
	l, err := net.Listen("tcp", connectionString)
	if err != nil {
		log.Fatal("Error listening: ", err.Error())
	}

	defer l.Close()

	log.Print("Listening on ", connectionString)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Error accepting: ", err.Error())
			continue
		}

		log.Println("Accepted connection from: ", conn.RemoteAddr())
		go s.handleRequest(conn)
	}
}

func (s *Server) createDirs(rootDir string) {
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		os.Mkdir(rootDir, 0777) // TODO: figure out right permissions
	}
	s.rootDir = rootDir

	objectsDir := filepath.Join(s.rootDir, "objects")
	if _, err := os.Stat(objectsDir); os.IsNotExist(err) {
		os.Mkdir(objectsDir, 0777)
	}
	s.objectsDir = objectsDir

	draftsDir := filepath.Join(s.rootDir, "drafts")
	if _, err := os.Stat(draftsDir); os.IsNotExist(err) {
		os.Mkdir(draftsDir, 0777)
	}
	s.draftsDir = draftsDir

	labelsDir := filepath.Join(s.rootDir, "labels")
	if _, err := os.Stat(labelsDir); os.IsNotExist(err) {
		os.Mkdir(labelsDir, 0777)
	}
	s.labelsDir = labelsDir
}

func (s *Server) handleRequest(conn net.Conn) {
	defer func() {
		log.Println("Closing the connection from: ", conn.RemoteAddr())
		conn.Close()
	}()

	// Make a buffer to hold incoming data (opcode + hash)
	buf := make([]byte, 1+32)

	for {
		opcode := buf[:1]
		_, err := conn.Read(opcode)
		if err != nil {
			log.Println("Error reading: ", err.Error())
			return
		}

		log.Println("Reading header")
		hash := buf[1:33]
		_, err = conn.Read(hash)
		if err != nil {
			log.Println("Error reading: ", err.Error())
			return
		}

		switch opcode[0] {
		case OP_SET:
			s.handleSetCommand(conn, hash)
		case OP_GET:
			s.handleGetCommand(conn, hash)
		case OP_EXISTS:
			log.Printf("Checking if hash %x exists\n", hash)
		case OP_SET_LABEL:
			log.Printf("Setting label\n")
			s.handleSetLabelCommand(conn, hash)
		default:
			log.Printf("Unknown command\n")
		}
	}
}

func (s *Server) handleSetCommand(conn net.Conn, hash []byte) {
	log.Printf("Setting hash: %x\n", hash)

	//TODO: create temp file which will hold the received data
	objectDirPath := filepath.Join(s.objectsDir, fmt.Sprintf("%x/%x", hash[:1], hash[1:2]))
	os.MkdirAll(objectDirPath, 0777)
	objectFilePath := filepath.Join(objectDirPath, fmt.Sprintf("%x", hash[2:]))

	f, err := os.Create(objectFilePath)
	if err != nil {
		log.Println("Error creating ", objectFilePath, err.Error())
		return
	}
	defer f.Close()

	var size int64
	binary.Read(conn, binary.LittleEndian, &size)
	log.Printf("Going to read %d bytes of file\n", size)

	for size > int64(0) {
		n := int64(1024)
		if size < int64(1024) {
			n = size
		}

		written, _ := io.CopyN(f, conn, n)
		size -= written
	}
}

func (s *Server) handleSetLabelCommand(conn net.Conn, hash []byte) {
	var size int64
	binary.Read(conn, binary.LittleEndian, &size)
	log.Printf("Going to read %d bytes of label\n", size)

	buf := make([]byte, size)
	_, err := conn.Read(buf)
	if err != nil {
		log.Println("Error reading label: ", err.Error())
	}

	log.Printf("Read label: %s (len=%d)\n", buf[:size], size)
	labelFilePath := filepath.Join(s.labelsDir, fmt.Sprintf("%s", buf[:size]))
	log.Println("Label file path: ", labelFilePath)
	f, err := os.Create(labelFilePath)
	if err != nil {
		log.Println("Error creating label file: ", err.Error())
		return
	}
	f.Write([]byte(fmt.Sprintf("%x", hash)))
	f.Close()
}

func (s *Server) handleGetCommand(conn net.Conn, hash []byte) {
	log.Printf("Getting hash: %x\n", hash)
}
