package main

import (
	"fmt"
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
	// DEBUG: test
	cmd = make([]string, 2, 2)
	cmd[0] = "/usr/bin/touch"
	cmd[1] = "/tmp/test-daemon"
	_, err = daemon.Daemon(0, 0, cmd)
	if err != nil {
		log.Println(err)
	}

	log.Println("DEBUG: 1")

	flagParse()
	fmt.Println("src:", mntFlags.src)
	fmt.Println("des:", mntFlags.des)
	fmt.Println("opt:", mntFlags.opt)

	return
}
