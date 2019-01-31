package main

import (
	"log"

	"github.com/gileshuang/multifs-fuse/daemon"
)

var (
	mntFlags mountFlags
)

func main() {
	var (
		err error
		cmd []string
	)
	flagParse()
	cmd = optToCmd()
	_, err = daemon.Daemon(0, 0, cmd)
	if err != nil {
		log.Println(err)
	}

	return
}
