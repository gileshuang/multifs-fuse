// Hellofs implements a simple "hello world" file system.
package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
	_ "bazil.org/fuse/fs/fstestutil"
	"golang.org/x/net/context"
	"github.com/alienhjy/multifs-fuse/daemon"
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "  %s MOUNTPOINT\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() != 1 {
		usage()
		os.Exit(2)
	}
	var newArgs []string
	newArgs = make([]string, 0, 3)
	newArgs = append(newArgs, os.Args[0])
	abspath, _ := filepath.Abs(os.Args[1])
	newArgs = append(newArgs, abspath)

	logFile, _ := filepath.Abs("multifs-debug.log")
	loger, logErr := os.OpenFile(logFile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	if logErr != nil {
		fmt.Println(logErr)
	}
	defer loger.Close()

	fmt.Fprintf(loger, "DEBUG: pre_daemon...\n")
	fmt.Fprintln(loger, "DEBUG: os.Args1:", os.Args)
	//daemon(1, 1)
	daemon.Daemon(0, 0, newArgs)
	fmt.Fprintf(loger, "DEBUG: post_daemon...\n")
	fmt.Fprintln(loger, "DEBUG: os.Args2:", os.Args)

	mountpoint := flag.Arg(0)
	fmt.Fprintln(loger, "DEBUG: mountpoint:", mountpoint)

	fmt.Fprintf(loger, "DEBUG: pre_fuse mount...\n")
	c, err := fuse.Mount(
		mountpoint,
		fuse.FSName("helloworld"),
		fuse.Subtype("hellofs"),
		fuse.LocalVolume(),
		fuse.VolumeName("Hello world!"),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close()
	fmt.Fprintf(loger, "DEBUG: poet_fuse mount...\n")

	fmt.Fprintf(loger, "DEBUG: pre_fs serve...\n")
	err = fs.Serve(c, FS{})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprintf(loger, "DEBUG: post_fs serve...\n")

	// check if the mount process has an error to report
	<-c.Ready
	if err := c.MountError; err != nil {
		log.Fatal(err)
	}
}

// FS implements the hello world file system.
type FS struct{}

func (FS) Root() (fs.Node, error) {
	return Dir{}, nil
}

// Dir implements both Node and Handle for the root directory.
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

// File implements both Node and Handle for the hello file.
type File struct{}

const greeting = "hello, world\n"

func (File) Attr(ctx context.Context, a *fuse.Attr) error {
	a.Inode = 2
	a.Mode = 0444
	a.Size = uint64(len(greeting))
	return nil
}

func (File) ReadAll(ctx context.Context) ([]byte, error) {
	return []byte(greeting), nil
}
