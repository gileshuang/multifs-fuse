package main

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"syscall"

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
	var (
		p    Dir
		pTmp fs.Node
		err  error
	)
	err = nil
	if dir.Path == "/" {
		a.Inode = 1
		if fusefs.readOnly == true {
			a.Mode = os.ModeDir | 0555
		} else {
			a.Mode = os.ModeDir | 0775
		}
	} else {
		pTmp, err = dir.Lookup(ctx, "..")
		if err != nil {
			return err
		}
		p = pTmp.(Dir)
		pFInfo, _ := os.Stat(filepath.Join(fusefs.target, p.Path))
		sysStat, ok := pFInfo.Sys().(*syscall.Stat_t)
		if !ok {
			return errors.New("Not a syscall.Stat_t")
		}
		a.Inode = fs.GenerateDynamicInode(sysStat.Ino,
			filepath.Base(dir.Path))
		// TODO: get file mode from backend
		if fusefs.readOnly == true {
			a.Mode = os.ModeDir | 0555
		} else {
			a.Mode = os.ModeDir | 0775
		}
	}
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

func (dir Dir) getDirent(fInfo os.FileInfo) (fuse.Dirent, error) {
	var (
		dirent fuse.Dirent
	)

	return dirent, nil
}

// ReadDirAll function read the all entry of this directory.
func (dir Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	var (
		dirmap  map[string]fuse.Dirent
		dirents []fuse.Dirent
	)
	log.Println("ReadDirAll", dir.Path)

	dirmap = make(map[string]fuse.Dirent)

	// Read from master
	mEntInfos, err := ioutil.
		ReadDir(filepath.Join(fusefs.master, dir.Path))
	if err != nil {
		return nil, err
	}
	for _, mEntInfo := range mEntInfos {
		entpath := filepath.Join(dir.Path, mEntInfo.Name())
		dirmap[entpath] = fuse.Dirent{Inode: 3,
			Name: mEntInfo.Name(),
			Type: fuse.DT_File}
	}

	// Read from slaves
	for _, slave := range fusefs.slaves {
		sEntInfos, err := ioutil.
			ReadDir(filepath.Join(slave, dir.Path))
		if err != nil {
			return nil, err
		}
	GetSlaves:
		for _, sEntInfo := range sEntInfos {
			entpath := filepath.Join(dir.Path, sEntInfo.Name())
			if _, ok := dirmap[entpath]; ok {
				continue GetSlaves
			}
			dirmap[entpath] = fuse.Dirent{Inode: 3,
				Name: sEntInfo.Name(),
				Type: fuse.DT_File}
		}
	}

	for _, ent := range dirmap {
		dirents = append(dirents, ent)
	}

	return dirents, nil
}
