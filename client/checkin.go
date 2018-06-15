package client

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"

	"github.com/adirg/gosmosis/server"
)

type Task struct {
	file string
	info os.FileInfo
	hash []byte
}

func Checkin(host string, port uint, dir string, label string) {
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
	filesToDigest := make(chan Task)
	filesToUpload := make(chan Task)
	filesToManifest := make(chan Task)

	wg.Add(1)
	go manifest(host, port, filesToManifest, label, &wg)

	wg.Add(1)
	go digest(filesToDigest, filesToUpload, filesToManifest, &wg)

	wg.Add(1)
	go upload(host, port, filesToUpload, &wg)

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		filesToDigest <- Task{path, info, []byte{}}
		return nil
	})

	if err != nil {
		log.Fatal("Error visiting ", absDir)
	}

	close(filesToDigest)
	wg.Wait()
}

func manifest(host string, port uint, filesToManifest chan Task, label string, wg *sync.WaitGroup) {
	defer wg.Done()

	connectionString := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		log.Fatal("Error dialing: ", err.Error())
	}

	defer func() {
		log.Println("Closing the connection")
		conn.Close()
	}()

	manifest := make(map[string]string)
	for {
		task, more := <-filesToManifest
		if !more {
			log.Println("Received all manifest entries")
			break
		}

		log.Printf("Adding %s to manifest\n", task.file)
		if len(task.hash) > 0 {
			manifest[task.file] = fmt.Sprintf("%x", task.hash)
		} else {
			manifest[task.file] = "nohash"
		}
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		log.Println("Error serializing manifest: ", err.Error())
	}
	fmt.Printf("Manifest: %s\n", manifestJSON)

	hash := sha256.Sum256(manifestJSON)

	conn.Write([]byte{server.OP_SET}) // Opcode
	conn.Write(hash[:])               // hash

	sizeBuf := make([]byte, 8)
	binary.PutVarint(sizeBuf, int64(len(manifestJSON)))
	log.Printf("Encoded size (%d): %v\n", len(sizeBuf), sizeBuf)
	log.Printf("Going to upload %d bytes of file\n", int64(len(manifestJSON)))
	binary.Write(conn, binary.LittleEndian, int64(len(manifestJSON)))

	conn.Write(manifestJSON)

	// set label
	conn.Write([]byte{server.OP_SET_LABEL}) // Opcode
	conn.Write(hash[:])                     // hash

	labelBuf := []byte(label)
	binary.PutVarint(sizeBuf, int64(len(labelBuf)))
	log.Printf("Encoded label size (%d): %v\n", len(sizeBuf), sizeBuf)
	log.Printf("Going to upload %d bytes of label\n", int64(len(labelBuf)))
	binary.Write(conn, binary.LittleEndian, int64(len(labelBuf)))

	conn.Write(labelBuf)
}

func digest(filesToDigest chan Task, filesToUpload chan Task, filesToManifest chan Task, wg *sync.WaitGroup) {
	defer wg.Done()

	for {
		task, more := <-filesToDigest
		if !more {
			log.Println("Received all files to digest")
			close(filesToUpload)
			close(filesToManifest)
			return
		}

		mode := task.info.Mode()
		if mode.IsDir() {
			log.Printf("Not going to digest %s (mode: %s)\n", task.file, mode)
			filesToManifest <- task
			continue
		}

		f, err := os.Open(task.file)
		if err != nil {
			log.Println("Error opening ", task.file)
			continue
		}
		defer f.Close()

		h := sha256.New()
		_, err = io.Copy(h, f)
		if err != nil {
			log.Println("Error digesting ", task.file)
			continue
		}

		task.hash = h.Sum(nil)
		log.Printf("sha256(%s) = %x\n", task.file, task.hash)
		filesToManifest <- task
		filesToUpload <- task
	}
}

func upload(host string, port uint, filesToUpload chan Task, wg *sync.WaitGroup) {
	defer wg.Done()

	connectionString := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		log.Fatal("Error dialing: ", err.Error())
	}

	defer func() {
		log.Println("Closing the connection")
		conn.Close()
	}()

	for {
		task, more := <-filesToUpload
		if !more {
			log.Println("Received all files to upload")
			return
		}

		log.Printf("Uploading %s (%x) \n", task.file, task.hash)
		conn.Write([]byte{0}) // Opcode
		conn.Write(task.hash) // hash

		info, err := os.Stat(task.file)
		size := info.Size()

		sizeBuf := make([]byte, 8)
		binary.PutVarint(sizeBuf, size)
		log.Printf("Encoded size (%d): %v\n", len(sizeBuf), sizeBuf)

		log.Printf("Going to upload %d bytes of file\n", size)
		binary.Write(conn, binary.LittleEndian, size)

		buf := make([]byte, 1024)
		f, err := os.Open(task.file)
		if err != nil {
			log.Println("Error opening ", task.file)
		}

		io.CopyBuffer(conn, f, buf)
	}
}
