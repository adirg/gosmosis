package client

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
)

type Task struct {
	file string
	info os.FileInfo
	hash []byte
}

func (c *Client) Checkin(workDir string, label string) {
	c.setWorkDir(workDir)

	filesToDigest := make(chan Task)
	filesToUpload := make(chan Task)
	filesToManifest := make(chan Task)

	c.wg.Add(1)
	go c.manifest(filesToManifest, label)

	c.wg.Add(1)
	go c.digest(filesToDigest, filesToUpload, filesToManifest)

	c.wg.Add(1)
	go c.upload(filesToUpload)

	err := filepath.Walk(c.workDir, func(path string, info os.FileInfo, err error) error {
		filesToDigest <- Task{path, info, []byte{}}
		return nil
	})

	if err != nil {
		log.Fatal("Error visiting ", c.workDir)
	}

	close(filesToDigest)
	c.wg.Wait()
}

func (c *Client) manifest(filesToManifest chan Task, label string) {
	defer c.wg.Done()

	conn := c.connect()
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

		rel, err := filepath.Rel(c.workDir, task.file)
		if err != nil {
			log.Println("Failed to get relative path of ", task.file)
			continue
		}

		log.Printf("Adding %s to manifest\n", rel)
		if len(task.hash) > 0 {
			manifest[rel] = fmt.Sprintf("%x", task.hash)
		} else {
			manifest[rel] = "nohash"
		}
	}

	manifestJSON, err := json.Marshal(manifest)
	if err != nil {
		log.Println("Error serializing manifest: ", err.Error())
	}
	fmt.Printf("Manifest: %s\n", manifestJSON)

	hash := sha256.Sum256(manifestJSON)
	size := int64(len(manifestJSON))
	r := bytes.NewReader(manifestJSON)
	set(conn, r, size, hash[:])
	setLabel(conn, label, hash[:])
}

func (c *Client) digest(filesToDigest chan Task, filesToUpload chan Task,
	filesToManifest chan Task) {
	defer c.wg.Done()

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

func (c *Client) upload(filesToUpload chan Task) {
	defer c.wg.Done()

	conn := c.connect()
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

		info, err := os.Stat(task.file)
		size := info.Size()

		f, err := os.Open(task.file)
		if err != nil {
			log.Println(err.Error())
		}

		set(conn, f, size, task.hash)
	}
}
