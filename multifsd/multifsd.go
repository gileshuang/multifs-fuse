package main

import (
	"fmt"
	"log"

	"bazil.org/fuse"
	"bazil.org/fuse/fs"
)

const (
	fsType = "multifs"
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
		fuse.Subtype(fsType),
		fuse.LocalVolume(),
		fuse.VolumeName("MultiFS"),
		fuse.MaxReadahead(fusefs.readSize),
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
