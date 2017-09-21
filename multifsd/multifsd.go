package main

import (
	"fmt"
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

const (
	// FsType is a const which named this filesystem type
	FsType = "multifs"
	// deletedMark is a string. Symlink in "master" who point
	// to deletedMark show the is no file in "Symlink", even if
	// there is a file exist in "slaves".
	deletedMark = "/dev/deleted"
)

var (
	fusefs FS
)

func main() {
	optsErr := flagParse()
	if optsErr != nil {
		log.Fatal(optsErr)
	}

	fmt.Println(fusefs)

	conn, err := fuse.Mount(
		fusefs.target,
		fuse.FSName(fusefs.master),
		fuse.Subtype(FsType),
		fuse.LocalVolume(),
		fuse.VolumeName("MultiFS"),
		fuse.MaxReadahead(fusefs.readSize),
		fuse.AllowDev(),
	)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	err = fs.Serve(conn, fusefs)
	if err != nil {
		log.Fatal(err)
	}

	// Check if the mount process has an error to report.
	<-conn.Ready
	if err := conn.MountError; err != nil {
		log.Fatal(err)
	}

	return
}
