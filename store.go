package main

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

const defaultRootFolderName = "ggnetwork"

func CASPathTransformFunc(key string) PathKey {
	hash := sha1.Sum([]byte(key))
	hashStr := hex.EncodeToString(hash[:])

	blockSize := 5
	sliceLen := len(hashStr) / blockSize
	paths := make([]string, sliceLen)

	for i := 0; i < sliceLen; i++ {
		from, to := i*blockSize, (i*blockSize)+blockSize
		paths[i] = hashStr[from:to]
	}

	return PathKey{
		PathName: strings.Join(paths, "/"),
		Filename: hashStr,
	}
}

type PathKey struct {
	PathName string
	Filename string
}

func (p PathKey) FullPath() string {
	return fmt.Sprintf("%s/%s", p.PathName, p.Filename)
}

func (p PathKey) FirstPathName() string {
	paths := strings.Split(p.PathName, "/")
	if len(paths) == 0 {
		return ""
	}
	return paths[0]
}

type PathTransformFunc func(string) PathKey

var DefaultPathTransformFunc = func(key string) PathKey {
	return PathKey{
		PathName: key,
		Filename: key,
	}
}

type StoreOpts struct {
	// Root is the folder name of the root, containing all the folders/files of the system.
	Root              string
	PathTransformFunc PathTransformFunc
}

type Store struct {
	StoreOpts StoreOpts
}

func NewStore(opts StoreOpts) *Store {
	if opts.PathTransformFunc == nil {
		opts.PathTransformFunc = DefaultPathTransformFunc
	}
	if len(opts.Root) == 0 {
		opts.Root = defaultRootFolderName
	}
	return &Store{
		StoreOpts: opts,
	}
}

func (s *Store) Has(key string) bool {
	pathKey := s.StoreOpts.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathKey.FullPath())

	_, err := os.Stat(fullPathWithRoot)
	return !errors.Is(err, os.ErrNotExist)
}

func (s *Store) Delete(key string) error {
	pathKey := s.StoreOpts.PathTransformFunc(key)

	defer func() {
		log.Printf("deleted [%s] from disk", pathKey.Filename)
	}()

	firstPathNameWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathKey.FirstPathName())

	return os.RemoveAll(firstPathNameWithRoot)
}

func (s *Store) Read(key string) (io.Reader, error) {
	f, err := s.ReadStream(key)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, f)

	return buf, nil
}

func (s *Store) ReadStream(key string) (io.ReadCloser, error) {
	pathKey := s.StoreOpts.PathTransformFunc(key)
	fullPathWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathKey.FullPath())
	return os.Open(fullPathWithRoot)
}

func (s *Store) WriteStream(key string, r io.Reader) error {
	pathKey := s.StoreOpts.PathTransformFunc(key)
	pathKeyWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathKey.PathName)
	if err := os.MkdirAll(pathKeyWithRoot, os.ModePerm); err != nil {
		return err
	}

	fullPathWithRoot := fmt.Sprintf("%s/%s", s.StoreOpts.Root, pathKey.FullPath())

	f, err := os.Create(fullPathWithRoot)
	if err != nil {
		return err
	}

	n, err := io.Copy(f, r)
	if err != nil {
		return err
	}

	log.Printf("written (%d) bytes to disk: %s", n, fullPathWithRoot)

	return nil
}
