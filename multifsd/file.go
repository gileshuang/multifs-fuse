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
	file *os.File
}

const greeting = "hello, world\n"

// Attr for get attr of file.
func (fl File) Attr(ctx context.Context, a *fuse.Attr) error {
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
		a.Mode = 0664
	}

	// Get file size.
	a.Size = uint64(fInfo.Size())
	return nil
}

// GetFullPath return the real path of File in backend.
func (fl File) GetFullPath() (string, error) {
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
func (fl File) Open(ctx context.Context, req *fuse.OpenRequest, resp *fuse.OpenResponse) (fs.Handle, error) {
	fullPath, err := fl.GetFullPath()
	if err != nil {
		fullPath = filepath.Join(fusefs.master, fl.Path)
		err = os.MkdirAll(filepath.Dir(fullPath), 0775)
		if err != nil {
			return nil, err
		}
	}
	fl.file, err = os.OpenFile(fullPath, int(req.Flags), 0666)
	if err != nil {
		return nil, err
	}

	resp.Flags |= fuse.OpenKeepCache

	return &fl, nil
}

// Release and close file
func (fl File) Release(ctx context.Context, req *fuse.ReleaseRequest) error {
	return fl.file.Close()
}

// Read function handle the read-request of File.
func (fl File) Read(ctx context.Context, req *fuse.ReadRequest, resp *fuse.ReadResponse) error {
	log.Println("File: ReadRequestSize:", req.Size)
	resp.Data = make([]byte, req.Size)
	nByte, err := fl.file.ReadAt(resp.Data, req.Offset)
	if err == io.ErrUnexpectedEOF || err == io.EOF {
		err = nil
	}

	log.Println("File: Read nByte:", nByte)

	return err
}

// ReadAll function read all of the file into []byte.
//func (fl File) ReadAll(ctx context.Context) ([]byte, error) {
//	fullPath, err := fl.GetFullPath()
//	if err != nil {
//		return nil, err
//	}
//
//	return ioutil.ReadFile(fullPath)
//}
