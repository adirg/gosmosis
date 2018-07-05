package client

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"path"
	"path/filepath"
)

func (c *Client) Checkout(dir string, label string) {
	c.setWorkDir(dir)

	filesToDownload := make(chan Task)

	c.wg.Add(1)
	go c.getManifest(filesToDownload, label)

	c.wg.Add(1)
	go c.download(filesToDownload, dir)

	c.wg.Wait()
}

func (c *Client) getManifest(filesToDownload chan Task, label string) {
	defer c.wg.Done()
	defer close(filesToDownload)

	conn := c.connect()
	defer func() {
		log.Println("Closing the connection")
		conn.Close()
	}()

	hash := getLabel(conn, label)

	var b bytes.Buffer
	get(conn, &b, hash)

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
	defer c.wg.Done()

	conn := c.connect()
	defer func() {
		log.Println("Closing the connection")
		conn.Close()
	}()

	for {
		task, more := <-filesToDownload
		if !more {
			log.Println("Downloaded all manifest entries")
			break
		}

		log.Printf("Downloading %s to manifest\n", task.file)
		absFile := path.Join(c.workDir, task.file)
		absDir := filepath.Dir(absFile)
		os.MkdirAll(absDir, 0777)
		f, err := os.OpenFile(absFile, os.O_WRONLY|os.O_CREATE, 0777)
		if err != nil {
			log.Println(err.Error())
		}

		get(conn, f, task.hash)
		f.Close()
	}
}
