package main

import (
	"io"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	"golang.org/x/net/context"
)

// File implements both Node and Handle for the files.
type File struct {
	Node
	//parent *Dir
	//mu     sync.RWMutex
	file *os.File
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

// Fsync file into backend file system.
func (fl *File) Fsync(ctx context.Context, req *fuse.FsyncRequest) error {
	log.Println("File: Fsync:", fl.Path)
	if fl.file == nil {
		// File is not opened
		return fuse.ENOTSUP
	}
	return fl.file.Sync()
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

	log.Println("File: Write nByte:", nByte)

	return err
}

// Readlink reads a symbolic link
//func (fl *File) Readlink(ctx context.Context, req *fuse.ReadlinkRequest) (string, error) {
//	fullPath, err := fl.getFullPath()
//	if err != nil {
//		return "", err
//	}
//	return os.Readlink(fullPath)
//}

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
