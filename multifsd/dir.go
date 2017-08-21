package main

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// Dir implements both Node and handle for the directorys.
type Dir struct {
	Path string
}

// Attr for get attr of directory.
func (dir Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 1
	a.Mode = os.ModeDir | 0555
	return nil
}

// Lookup return the sub Node in this directory.
func (dir Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	if name == "hello" {
		return File{}, nil
	}
	return nil, fuse.ENOENT
}

// ReadDirAll function read the all entry of this directory.
func (dir Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var (
		dirents []fuse.Dirent
	)
	log.Println("ReadDirAll", dir.Path)
	mEntInfos, err := ioutil.
		ReadDir(filepath.Join(fusefs.master, dir.Path))
	if err != nil {
		return nil, err
	}
	for _, mEntInfo := range mEntInfos {
		dirents = append(dirents,
			fuse.Dirent{Inode: 3,
				Name: mEntInfo.Name(),
				Type: fuse.DT_File})
	}

	return dirents, nil
}
