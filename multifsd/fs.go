package main

import (
	"log"

	"bazil.org/fuse/fs"
)

// FS implements the file system.
type FS struct {
	// Where this fuse filesystem mounted.
	target string
	// Where we read files from first.
	master string
	// If read from master failed, where we read files from.
	slaves strSlice
	// If we should copy file from slaves to master.
	copyOnRead bool
	// If filesystem mounted as readonly
	readOnly bool
}

// Root implement the ROOT of filesystem.
func (FS) Root() (fs.Node, error) {
	var dir = Dir{Path: "/"}
	log.Println("Set root.")
	return dir, nil
}
