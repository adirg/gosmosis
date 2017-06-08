package main

import "fmt"
import "log"
import "github.com/adirg/gosmosis/store"

func main() {
	store, err := store.NewStore("/tmp/bla")
	if err != nil {
		log.Fatal(err)
	}
	if store.Exist([]byte("aabb")) {
		fmt.Println("key does exist")
	} else {
		fmt.Println("key doesn't exist")
	}
}
