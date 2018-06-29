package client

import (
	"log"
	"os"
	"path/filepath"
	"sync"
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
