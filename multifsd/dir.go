package main

import (
    "os"

    "bazil.org/fuse"
    "bazil.org/fuse/fs"
    "golang.org/x/net/context"
)

// Dir implements both Node and handle for the directorys.
type Dir struct{}

func (Dir) Attr(ctx context.Context, a *fuse.Attr) error {
    a.Inode = 1
    a.Mode = os.ModeDir | 0555
    return nil
}

func (Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
    if name == "hello" {
        return File{}, nil
    }
    return nil, fuse.ENOENT
}

var dirDirs = []fuse.Dirent{
    {Inode: 2, Name: "hello", Type: fuse.DT_File},
}

func (Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
    return dirDirs, nil
}

