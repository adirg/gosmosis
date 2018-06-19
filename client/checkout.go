package client

import (
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/adirg/gosmosis/server"
)

func Checkout(host string, port uint, dir string, label string) {
	absDir, err := filepath.Abs(dir)
	if err != nil {
		log.Fatal("Error determining absolute path of directory: ", err.Error())
	}

	mode, err := os.Stat(absDir)
	if err != nil {
		log.Fatal(err.Error())
	}

	if !mode.IsDir() {
		log.Fatalf("%s is not a directory", absDir)
	}

	var wg sync.WaitGroup
	filesToDownload := make(chan Task)

	wg.Add(1)
	go getManifest(host, port, filesToDownload, label, &wg)

	wg.Add(1)
	go download(&wg)

	wg.Wait()
}

func getManifest(host string, port uint, filesToDownload chan Task, label string, wg *sync.WaitGroup) {
	defer wg.Done()
	defer close(filesToDownload)

	connectionString := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		log.Fatal("Error dialing: ", err.Error())
	}

	defer func() {
		log.Println("Closing the connection")
		conn.Close()
	}()

	conn.Write([]byte{server.OpGetLabel})

	sizeBuf := make([]byte, 8)
	labelBuf := []byte(label)
	binary.PutVarint(sizeBuf, int64(len(labelBuf)))
	log.Printf("Encoded label size (%d): %v\n", len(sizeBuf), sizeBuf)
	log.Printf("Going to upload %d bytes of label\n", int64(len(labelBuf)))
	binary.Write(conn, binary.LittleEndian, int64(len(labelBuf)))

	// get manifest hash
	hash := make([]byte, 32)
	_, err = io.ReadFull(conn, hash)
	if err != nil {
		log.Println("Failed to read hash of label: ", label)
		return
	}

	// get manifest content
	manifestFile, err := ioutil.TempFile("", "manifest")
	if err != nil {
		log.Println("Failed to open temporary manifest file")
		return
	}

	defer manifestFile.Close()

	conn.Write([]byte{server.OpGet})
	conn.Write(hash)

	var size int64
	binary.Read(conn, binary.LittleEndian, &size)
	log.Println("Size of manifest file: ", size)

	r := io.LimitReader(conn, size)
	buf := make([]byte, 1024)
	_, err = io.CopyBuffer(manifestFile, r, buf)
	if err != nil {
		log.Println("Failed to read manifest content")
		return
	}
}

func download(wg *sync.WaitGroup) {
	defer wg.Done()
	fmt.Println("Not implmented yet")
	//connectionString := fmt.Sprintf("%s:%d", host, port)
	//conn, err := net.Dial("tcp", connectionString)
	//if err != nil {
	//log.Fatal("Error dialing: ", err.Error())
	//}

	//defer func() {
	//log.Println("Closing the connection")
	//conn.Close()
	//}()

	//for {
	//task, more := <-filesToDownload
	//if !more {
	//log.Println("Received all files to download")
	//return
	//}

	//log.Printf("downloading %s (%x) \n", task.file, task.hash)
	//conn.Write([]byte{server.OpGet}) // Opcode
	//conn.Write(task.hash)            // hash

	//info, err := os.Stat(task.file)
	//size := info.Size()

	//sizeBuf := make([]byte, 8)
	//binary.PutVarint(sizeBuf, size)
	//log.Printf("Encoded size (%d): %v\n", len(sizeBuf), sizeBuf)

	//log.Printf("Going to upload %d bytes of file\n", size)
	//binary.Write(conn, binary.LittleEndian, size)

	//buf := make([]byte, 1024)
	//f, err := os.Open(task.file)
	//if err != nil {
	//log.Println("Error opening ", task.file)
	//}

	//io.CopyBuffer(conn, f, buf)
	//}
}
