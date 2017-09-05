package main

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"syscall"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// File implements both Node and Handle for the files.
type File struct {
	Path string
	//parent *Dir
	//mu     sync.RWMutex
	file *os.File
}

// Attr for get attr of file.
func (fl *File) Attr(ctx context.Context, a *fuse.Attr) error {
	log.Println("File: Attr:", fl.Path)
	pFInfo, err := os.Stat(filepath.Join(fusefs.target, fl.Path, ".."))
	if err != nil {
		return err
	}
	fInfo, err := os.Lstat(filepath.Join(fusefs.master, fl.Path))
	if err != nil {
	GetSlaves:
		for _, slave := range fusefs.slaves {
			fInfo, err = os.Lstat(filepath.Join(slave, fl.Path))
			if err != nil {
				continue GetSlaves
			}
			break GetSlaves
		}
	}

	// Get file inode.
	sysStat, ok := pFInfo.Sys().(*syscall.Stat_t)
	if !ok {
		return errors.New("Not a syscall.Stat_t")
	}
	a.Inode = fs.GenerateDynamicInode(sysStat.Ino,
		filepath.Base(fl.Path))

	// TODO: Get file mode from backend.
	if fusefs.readOnly == true {
		a.Mode = 0444
	} else {
		a.Mode = fusefs.defFileMode
	}

	// Get file size.
	a.Size = uint64(fInfo.Size())
	blocks := a.Size / fusefs.unitSize
	if 0 != a.Size%fusefs.unitSize {
		blocks++
	}
	a.Blocks = blocks
	return nil
}

// GetFullPath return the real path of File in backend.
func (fl *File) getFullPath() (string, error) {
	fullPath := filepath.Join(fusefs.master, fl.Path)
	_, err := os.Lstat(fullPath)
	if err != nil {
	GetSlaves:
		for _, slave := range fusefs.slaves {
			fullPath = filepath.Join(slave, fl.Path)
			_, err = os.Lstat(fullPath)
			if err != nil {
				continue GetSlaves
			}
			break GetSlaves
		}
		if err != nil {
			fullPath = ""
		}
	}

	return fullPath, err
}

// Open file
func (fl *File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	log.Println("File: Open:", fl.Path)
	fullPath, err := fl.getFullPath()
	if err != nil {
		fullPath = filepath.Join(fusefs.master, fl.Path)
		err = os.MkdirAll(filepath.Dir(fullPath), fusefs.defDirMode)
		if err != nil {
			return nil, err
		}
	}
	fl.file, err = os.OpenFile(fullPath, int(req.Flags), fusefs.defFileMode)
	if err != nil {
		return nil, err
	}

	resp.Flags |= fuse.OpenKeepCache

	return fl, nil
}

// Release and close file
func (fl *File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	log.Println("File: Release:", fl.Path)
	return fl.file.Close()
}

// Read function handle the read-request of File.
func (fl *File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Println("File: ReadRequestSize:", req.Size)
	if fl.file == nil {
		// File is not opened
		return fuse.ENOTSUP
	}
	resp.Data = make([]byte, req.Size)
	nByte, err := fl.file.ReadAt(resp.Data, req.Offset)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}

	log.Println("File: Read nByte:", nByte)

	return err
}

// Write function handle the write-request of File.
func (fl *File) Write(ctx context.Context, req *fuse.WriteRequest, resp *fuse.WriteResponse) error {
	log.Println("File: Write:", fl.Path)
	if fl.file == nil {
		// File is not opened
		return fuse.ENOTSUP
	}
	nByte, err := fl.file.WriteAt(req.Data, req.Offset)
	resp.Size = nByte

	return err
}

// ReadAll function read all of the file into []byte.
//func (fl *File) ReadAll(ctx context.Context) ([]byte, error) {
//	log.Println("File: ReadAll:", fl.Path)
//	fullPath, err := fl.getFullPath()
//	if err != nil {
//		return nil, err
//	}
//
//	return ioutil.ReadFile(fullPath)
//}
