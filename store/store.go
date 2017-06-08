package store

import "os"
import "path"
import "fmt"
import "errors"

type Store struct {
	root string
}

func NewStore(root string) (*Store, error) {
	fmt.Println("Checking if dir exists: ", root)
	dirExists, _ := exists(root)
	fmt.Println(dirExists)
	if !dirExists {
		return nil, errors.New("Root direcotry doesn't exists")
	}
	store := new(Store)
	store.root = root
	return store, nil
}

func (store *Store) Set(key []byte, val []byte) error {
	return nil
}

func (store *Store) Get(key []byte) ([]byte, error) {
	return nil, nil
}

func (store *Store) Exist(key []byte) bool {
	keySize := len(key)
	if keySize <= 2 {
		return false
	}
	keyPath := path.Join(store.root, string(key[:2]), string(key[2:]))
	_, err := os.Stat(keyPath)
	if err != nil {
		return false
	}
	return true
}

// exists returns whether the given file or directory exists or not
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}
