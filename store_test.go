package main

import (
	"bytes"
	"io"
	"testing"
)

func TestPathTransformFunc(t *testing.T) {
	key := "mybestpicture"
	pathKey := CASPathTransformFunc(key)
	expectedOriginalKey := "be17b32c2870b1c0c73b59949db6a3be7814dd23"
	expectedPathName := "be17b/32c28/70b1c/0c73b/59949/db6a3/be781/4dd23"
	if pathKey.PathName != expectedPathName {
		t.Errorf("have %s want %s", pathKey.PathName, expectedPathName)
	}
	if pathKey.Filename != expectedOriginalKey {
		t.Errorf("have %s want %s", pathKey.Filename, expectedOriginalKey)
	}
}

func TestStore(t *testing.T) {
	s := newStore()
	defer teardown(t, s)

	key := "foobar"

	data := []byte("some jpg bytes")
	if err := s.Write(key, bytes.NewReader(data)); err != nil {
		t.Error(err)
	}

	if ok := s.Has(key); !ok {
		t.Errorf("expected to have key %s", key)
	}

	r, err := s.Read(key)
	if err != nil {
		t.Error(err)
	}

	b, _ := io.ReadAll(r)
	if string(b) != string(data) {
		t.Errorf("want %s have %s", data, b)
	}

	if err := s.Delete(key); err != nil {
		t.Error(err)
	}

	if ok := s.Has(key); ok {
		t.Errorf("expected to not have key %s", key)
	}
}

func newStore() *Store {
	opts := StoreOpts{
		PathTransformFunc: CASPathTransformFunc,
	}

	return NewStore(opts)
}

func teardown(t *testing.T, s *Store) {
	if err := s.Clear(); err != nil {
		t.Error(err)
	}
}
