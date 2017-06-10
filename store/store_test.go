package store

import "io/ioutil"
import "os"
import "testing"
import "github.com/stretchr/testify/assert"

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
	if store.Exist("a") {
		t.Error("Expected key to not exist")
	}
}

func TestSet(t *testing.T) {
	root, err := ioutil.TempDir("", "store")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(root)

	store, err := NewStore(root)
	if err != nil {
		t.Error(err)
	}

	err = store.Set("a", []byte("my value"))
	assert.Error(t, err)

	err = store.Set("aaaa", []byte("my value"))
	assert.Nil(t, err)
}

func TestGet(t *testing.T) {
	root, err := ioutil.TempDir("", "store")
	if err != nil {
		t.Error(err)
	}
	defer os.RemoveAll(root)

	store, err := NewStore(root)
	if err != nil {
		t.Error(err)
	}
	_, err = store.Get("aa")
	assert.Error(t, err)
	store.Set("aaaa", []byte("my value"))
	_, err = store.Get("aaaa")
	assert.Nil(t, err)
}
