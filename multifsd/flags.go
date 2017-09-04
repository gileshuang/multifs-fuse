package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
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

func flagParse() error {
	var (
		finfo os.FileInfo
		err   error
	)

	// Init vars
	fusefs.slaves = make(strSlice, 0, 4)

	// Parse
	flag.StringVar(&fusefs.target, "target", "", "Where this filesystem mounted, this value should not be empty.")
	flag.StringVar(&fusefs.master, "master", "", "The master backend, where read files first. This value should not be empty.")
	flag.Var(&fusefs.slaves, "slaves", "Optional. The slave backends of reading files.")
	flag.BoolVar(&fusefs.copyOnRead, "cor", false, "Copy on Read. Set it to true for enable copy-file-form-slave-to-master.")
	flag.BoolVar(&fusefs.readOnly, "ro", false, "ReadOnly. Set it to true for disable write-to-master.")
	flag.Parse()

	// Check flags
	if len(fusefs.master) == 0 || len(fusefs.target) == 0 {
		return errors.New("Both -target and -master should not be empty")
	}

	// Check if dirs exist.(target, master, slaves)
	finfo, err = os.Lstat(fusefs.target)
	if err != nil {
		return err
	} else if finfo.IsDir() == false {
		return errors.New("Target should be a directory")
	}
	finfo, err = os.Lstat(fusefs.master)
	if err != nil {
		return err
	} else if finfo.IsDir() == false {
		return errors.New("Master should be a directory")
	}
	for _, slave := range fusefs.slaves {
		finfo, err = os.Lstat(slave)
		if err != nil {
			return err
		} else if finfo.IsDir() == false {
			return errors.New("Target should be a directory")
		}
	}
	// Set size of units to default:512
	fusefs.unitSize = 512
	fusefs.readSize = 128 * 1024
	return nil
}
