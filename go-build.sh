#!/bin/bash

export GOPATH=`pwd`:${GOPATH}
export GOBIN=`pwd`/bin
export PATH=${GOBIN}:${PATH}
go install github.com/gileshuang/multifs-fuse/multifsd
go install github.com/gileshuang/multifs-fuse/mount.multifs
