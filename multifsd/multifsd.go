package main

import (
    "fmt"
    "log"

    "bazil.org/fuse"
    "bazil.org/fuse/fs"
)

const (
    FsType = "multifs"
)

func main() {
    optsErr := flagParse()
    if optsErr != nil {
        log.Fatal(optsErr)
    }

    fmt.Println(opts)

    conn, err := fuse.Mount(
        opts.target,
        fuse.FSName(opts.master),
        fuse.Subtype(FsType),
        fuse.LocalVolume(),
        fuse.VolumeName("MultiFS"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer conn.Close()

    err = fs.Serve(conn, FS{})
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

