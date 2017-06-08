package store

import "io/ioutil"
import "os"
import "testing"

func TestNewStore(t *testing.T) {
	root, err := ioutil.TempDir("", "store")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(root)

	_, err = NewStore(root)
	if err != nil {
		t.Error("Expected error due to missing root directory")
	}
}

func TestExist(t *testing.T) {
	root, err := ioutil.TempDir("", "store")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(root)

	store, err := NewStore(root)
	if err != nil {
		t.Error(err)
	}
	if store.Exist([]byte("a")) {
		t.Error("Expected key to not exist")
	}
}
