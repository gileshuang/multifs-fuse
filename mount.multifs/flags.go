package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
)

type strSlice []string

func (slf *strSlice) String() string {
	return fmt.Sprint(*slf)
}

func (slf *strSlice) Set(value string) error {
	for _, tmp := range *slf {
		if value == tmp {
			return errors.New("Duplicate flags")
		}
	}
	*slf = append(*slf, value)
	return nil
}

type mountFlags struct {
	src string
	des string
	opt strSlice
}

func optHelpText() string {
	var (
		text string
	)
	text = "mount option"
	return text
}

func flagParse() error {
	var (
		strFlags strSlice
		err      error
	)
	flag.Var(&strFlags, "o", optHelpText())
	flag.Parse()
	if len(flag.Args()) == 2 {
		mntFlags.src = flag.Arg(0)
		mntFlags.des = flag.Arg(1)
		mntFlags.opt = strFlags
	} else {
		flag.Usage()
		log.Println("len(flag.Args()):", len(flag.Args()))
	}

	return err
}
