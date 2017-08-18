package main

import (
    "fmt"
    "os"
    "flag"
    "errors"
)

type strSlice []string

func (slf *strSlice) String() string {
    return fmt.Sprint(*slf)
}

func (slf *strSlice) Set(value string) error {
    for _, tmp := range *slf {
        if value == tmp {
            return errors.New("Duplicate flags.")
        }
    }
    *slf = append(*slf, value)
    return nil
}

type options struct {
    // Where this fuse filesystem mounted.
    target string
    // Where we read files from first.
    master string
    // If read from master failed, where we read files from.
    slaves strSlice
    // If we should copy file from slaves to master.
    copyOnRead bool
}

var (
    opts options
)

func flagParse() error {
    var (
        finfo os.FileInfo
        err error
    )

    // Init vars
    opts.slaves = make(strSlice, 0, 4)

    // Parse
    flag.StringVar(&opts.target, "target", "", "Where this filesystem mounted, this value should not be empty.")
    flag.StringVar(&opts.master, "master", "", "The master backend, where read files first. This value should not be empty.")
    flag.Var(&opts.slaves, "slaves", "Optional. The slave backends of reading files.")
    flag.BoolVar(&opts.copyOnRead, "cor", false, "Copy on Read. Set it to true for enable copy-file-form-slave-to-master.")
    flag.Parse()

    // Check flags
    if len(opts.master) == 0 || len(opts.target) == 0 {
        return errors.New("Both -target and -master should not be empty.")
    }

    // Check if dirs exist.(target, master, slaves)
    finfo, err = os.Lstat(opts.target)
    if err != nil {
        return err
    } else if finfo.IsDir() == false {
        return errors.New("target should be a directory.")
    }
    finfo, err = os.Lstat(opts.master)
    if err != nil {
        return err
    } else if finfo.IsDir() == false {
        return errors.New("master should be a directory.")
    }
    for _, slave := range opts.slaves {
        finfo, err = os.Lstat(slave)
        if err != nil {
            return err
        } else if finfo.IsDir() == false {
            return errors.New("target should be a directory.")
        }
    }
    return nil
}

