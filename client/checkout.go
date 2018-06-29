package client

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/adirg/gosmosis/server"
)

func (c *Client) Checkout(dir string, label string) {
	c.setWorkDir(dir)

	filesToDownload := make(chan Task)
	go c.getManifest(filesToDownload, label)
	go c.download(filesToDownload, dir)

	c.wg.Wait()
}

func (c *Client) getManifest(filesToDownload chan Task, label string) {
	c.wg.Add(1)
	defer c.wg.Done()
	defer close(filesToDownload)

	connectionString := fmt.Sprintf("%s:%d", c.host, c.port)
	conn, err := net.Dial("tcp", connectionString)
	if err != nil {
		log.Fatal("Error dialing: ", err.Error())
	}

	defer func() {
		log.Println("Closing the connection")
		conn.Close()
	}()

	conn.Write([]byte{server.OpGetLabel})

	labelBuf := []byte(label)
	log.Printf("Going to download %d bytes of label\n", int64(len(labelBuf)))
	binary.Write(conn, binary.LittleEndian, int64(len(labelBuf)))

	// send label to the server
	conn.Write(labelBuf)

	// get manifest hash
	hash := make([]byte, 32)
	_, err = io.ReadFull(conn, hash)
	if err != nil {
		log.Println("Failed to read hash of label: ", label)
		return
	}

	log.Printf("label hash: %x\n", hash)

	conn.Write([]byte{server.OpGet})
	conn.Write(hash)

	var size int64
	binary.Read(conn, binary.LittleEndian, &size)
	log.Println("Size of manifest file: ", size)

	var b bytes.Buffer

	r := io.LimitReader(conn, size)
	buf := make([]byte, 1024)
	_, err = io.CopyBuffer(&b, r, buf)
	if err != nil {
		log.Println("Failed to read manifest content")
		return
	}

	var manifestObj map[string]string
	log.Println("manifest: ", b.String())
	json.Unmarshal(b.Bytes(), &manifestObj)

	for filename, hash := range manifestObj {
		log.Println("filename: ", filename)
		log.Println("hash:   : ", hash)
		if hash != "nohash" {
			hashBuf, _ := hex.DecodeString(string(hash))
			filesToDownload <- Task{filename, nil, hashBuf}
		}
	}
}

func (c *Client) download(filesToDownload chan Task, dir string) {
	c.wg.Add(1)
	defer c.wg.Done()

	for {
		task, more := <-filesToDownload
		if !more {
			log.Println("Downloaded all manifest entries")
			break
		}

		log.Printf("Downloading %s to manifest\n", task.file)
	}
}
