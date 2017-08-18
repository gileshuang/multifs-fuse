package main

import (
    "bazil.org/fuse/fs"
)

// FS implements the file system.
type FS struct{}

func (FS) Root() (fs.Node, error) {
    return Dir{}, nil
}

