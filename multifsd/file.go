package main

import (
	"bazil.org/fuse"
	"golang.org/x/net/context"
)

// File implements both Node and Handle for the files.
type File struct {
	Path string
}

const greeting = "hello, world\n"

// Attr for get attr of file.
func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(greeting))
	return nil
}

// ReadAll function read all of the file into []byte.
func (File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(greeting), nil
}
