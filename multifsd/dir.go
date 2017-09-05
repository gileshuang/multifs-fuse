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
func (dir *Dir) Attr(ctx context.Context, a *fuse.Attr) error {
	//log.Println("Dir: Attr:", dir.Path)
	var ()
	pFInfo, _ := os.Stat(filepath.Dir(filepath.Join(fusefs.target, dir.Path)))
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
		a.Mode = os.ModeDir | fusefs.defDirMode
	}
	return nil
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
			return &Dir{Path: nodePath}, nil
		}
		log.Println("Dir:", nodePath, "is file")
		return &File{Path: nodePath}, nil
	}

	// Lookup from slaves
	for _, slave := range fusefs.slaves {
		sFInfo, err := os.Lstat(filepath.Join(slave, nodePath))
		if err == nil {
			if sFInfo.IsDir() {
				log.Println("Dir:", nodePath, "is dir")
				return &Dir{Path: nodePath}, nil
			}
			log.Println("Dir:", nodePath, "is file")
			return &File{Path: nodePath}, nil
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

// Access checks wheather operation has permission
func (dir *Dir) Access(ctx context.Context, req *fuse.AccessRequest) error {
	// TODO: check permission
	return nil
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
	newFile = &File{Path: filepath.Join(dir.Path, req.Name)}
	newFile.file, err = os.Create(fullFilePath)
	if err != nil {
		return &File{}, &File{}, err
	}

	return newFile, newFile, nil
}
