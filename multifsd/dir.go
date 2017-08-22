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
	nodePath := filepath.Join(dir.Path, name)
	log.Println("Lookup", nodePath)

	// Lookup from master
	mFInfo, err := os.Lstat(filepath.Join(fusefs.master, nodePath))
	if err == nil {
		if mFInfo.IsDir() {
			log.Println(nodePath, "is dir")
			return Dir{Path: nodePath}, nil
		}
		log.Println(nodePath, "is file")
		return File{Path: nodePath}, nil
	}

	// Lookup from slaves
	for _, slave := range fusefs.slaves {
		sFInfo, err := os.Lstat(filepath.Join(slave, nodePath))
		if err == nil {
			if sFInfo.IsDir() {
				log.Println(nodePath, "is dir")
				return Dir{Path: nodePath}, nil
			}
			log.Println(nodePath, "is file")
			return File{Path: nodePath}, nil
		}
	}

	return nil, fuse.ENOENT
}

// ReadDirAll function read the all entry of this directory.
func (dir Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var (
		dirents []fuse.Dirent
	)
	log.Println("ReadDirAll", dir.Path)

	// Read from master
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

	// Read from slaves
	for _, slave := range fusefs.slaves {
		sEntInfos, err := ioutil.
			ReadDir(filepath.Join(slave, dir.Path))
		if err != nil {
			return nil, err
		}
		for _, sEntInfo := range sEntInfos {
			dirents = append(dirents,
				fuse.Dirent{Inode: 3,
					Name: sEntInfo.Name(),
					Type: fuse.DT_File})
		}
	}

	return dirents, nil
}
