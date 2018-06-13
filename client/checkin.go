package client

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sync"
)

type Task struct {
	file string
	info os.FileInfo
	hash []byte
}

func Checkin(host string, port uint, dir string) {
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
	go manifest(filesToManifest, &wg)

	wg.Add(1)
	go digest(filesToDigest, filesToUpload, filesToManifest, &wg)

	wg.Add(1)
	go upload(host, port, filesToUpload, &wg)

	err = filepath.Walk(absDir, func(path string, info os.FileInfo, err error) error {
		mode := info.Mode()
		if mode.IsDir() {
			log.Printf("Not going to digest %s (mode: %s)\n", path, mode)
			return nil
		}

		filesToDigest <- Task{path, info, []byte{}}
		return nil
	})

	if err != nil {
		log.Fatal("Error visiting ", absDir)
	}

	close(filesToDigest)
	wg.Wait()
}

func manifest(filesToManifest chan Task, wg *sync.WaitGroup) {
	defer wg.Done()

	manifest := make(map[string]string)
	for {
		task, more := <-filesToManifest
		if !more {
			log.Println("Received all manifest entries")
			return
		}

		manifest[task.file] = fmt.Sprintf("%x", task.hash)
	}
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
			return
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
