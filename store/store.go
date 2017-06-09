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

func (store *Store) Set(key string, val []byte) error {
	keySize := len(key)
	if keySize <= 2 {
		return errors.New("Invalid key")
	}

	keySubdir = path.Join(store.root, key[:2])
	subdirExists, _ := exists(keySubdir)
	if !subdirExists {
		err := os.Mkdir(keySubdir)
		if err != nil {
			return err
		}
	}

	keyFullPath = os.Join(keySubdir, key[2:])
	f, err := os.Create(keyFullPath)
	if err != nil {
		return err
	}

	n, err := f.Write(val)
	if err != nil {
		return err
	}

	return nil
}

func (store *Store) Get(key string) ([]byte, error) {
	return nil, nil
}

func (store *Store) Exist(key string) bool {
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
