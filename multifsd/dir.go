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
	Node
}

// Lookup return the sub Node in this directory.
func (dir *Dir) Lookup(ctx context.Context, name string) (fs.Node, error) {
	log.Println("Dir: Lookup:", name)
	nodePath := filepath.Join(dir.Path, name)

	// Lookup from master
	mFInfo, err := os.Lstat(filepath.Join(fusefs.master, nodePath))
	if err == nil {
		if mFInfo.IsDir() {
			log.Println("Dir:", nodePath, "is dir")
			return &Dir{Node: Node{Path: nodePath}}, nil
		}
		log.Println("Dir:", nodePath, "is file")
		return &File{Node: Node{Path: nodePath}}, nil
	}

	// Lookup from slaves
	for _, slave := range fusefs.slaves {
		sFInfo, err := os.Lstat(filepath.Join(slave, nodePath))
		if err == nil {
			if sFInfo.IsDir() {
				log.Println("Dir:", nodePath, "is dir")
				return &Dir{Node: Node{Path: nodePath}}, nil
			}
			log.Println("Dir:", nodePath, "is file")
			return &File{Node: Node{Path: nodePath}}, nil
		}
	}

	log.Println("Dir: Lookup: done")
	return nil, fuse.ENOENT
}

func (dir *Dir) getDirent(fInfo os.FileInfo) (fuse.Dirent, error) {
	var (
		dirent fuse.Dirent
	)

	return dirent, nil
}

// ReadDirAll function read the all entry of this directory.
func (dir *Dir) ReadDirAll(ctx context.Context) ([]fuse.Dirent, error) {
	log.Println("Dir: ReadDirAll:", dir.Path)
	var (
		dirmap     map[string]fuse.Dirent
		dirents    []fuse.Dirent
		dirAbsPath string
		allerr     error
	)

	dirmap = make(map[string]fuse.Dirent)
	dirAbsPath = filepath.Join(fusefs.target, dir.Path)

	// Add default entry
	dirmap[dirAbsPath] = fuse.Dirent{Name: "."}
	dirmap[filepath.Dir(dirAbsPath)] = fuse.Dirent{Name: ".."}

	// Read from master
	mEntInfos, err := ioutil.
		ReadDir(filepath.Join(fusefs.master, dir.Path))
	if err != nil {
		allerr = err
	}
	for _, mEntInfo := range mEntInfos {
		entpath := filepath.Join(dirAbsPath, mEntInfo.Name())
		dirmap[entpath] = fuse.Dirent{Name: mEntInfo.Name()}
	}

	// Read from slaves
	for _, slave := range fusefs.slaves {
		sEntInfos, err := ioutil.
			ReadDir(filepath.Join(slave, dir.Path))
		if err == nil {
			allerr = nil
		}
	GetSlaves:
		for _, sEntInfo := range sEntInfos {
			entpath := filepath.Join(dirAbsPath, sEntInfo.Name())
			if _, ok := dirmap[entpath]; ok {
				continue GetSlaves
			}
			dirmap[entpath] = fuse.Dirent{Name: sEntInfo.Name()}
		}
	}

	for _, ent := range dirmap {
		dirents = append(dirents, ent)
	}

	if allerr != nil {
		return nil, allerr
	}
	return dirents, nil
}

// Create a new directory entry
func (dir *Dir) Create(ctx context.Context, req *fuse.CreateRequest, resp *fuse.CreateResponse) (fs.Node, fs.Handle, error) {
	log.Println("Dir: Create:", req.Name)
	var (
		newFile *File
		err     error
	)
	newFilePath := filepath.Join(dir.Path, req.Name)
	fullFilePath := filepath.Join(fusefs.master, newFilePath)
	fullDirPath := filepath.Dir(fullFilePath)
	err = os.MkdirAll(fullDirPath, fusefs.defDirMode)
	if err != nil {
		return &File{}, &File{}, err
	}
	newFile = &File{Node: Node{Path: newFilePath}}
	newFile.file, err = os.Create(fullFilePath)
	if err != nil {
		return &File{}, &File{}, err
	}

	return newFile, newFile, nil
}

// Mkdir under this Dir
func (dir *Dir) Mkdir(ctx context.Context, req *fuse.MkdirRequest) (fs.Node, error) {
	var (
		newDir   *Dir
		err      error
		modePerm os.FileMode
		newMode  os.FileMode
	)
	newPath := filepath.Join(dir.Path, req.Name)
	fullNewPath := filepath.Join(fusefs.master, newPath)
	if req.Mode.IsDir() != true {
		modePerm = 0777 - req.Umask
	} else {
		modePerm = req.Mode.Perm()
	}
	newMode = req.Mode/01000 + modePerm
	err = os.MkdirAll(fullNewPath, newMode)
	newDir = new(Dir)
	newDir.Path = newPath

	return newDir, err
}
