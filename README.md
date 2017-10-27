# multifs-fuse
A fuse filesystem with multiple backend.

## Go Dependence
install dependence if build from source code.
```
go get -v -u bazil.org/fuse
go get -v -u golang.org/x/net/context
```

## Features

- Read dir from multiple backend.
- Read file from one of multiple backend.
- Mkdir on master backend.
- Get attr/stat of file and dir from backend.
- Create new file on master backend.
- Remove file or directory on master.
- Mark file or directory as *deleted* if this file is existed on slaves backend.
- Open file.
- Read file.
- Write new change to file, with locate on original backend.
- Sync file to disk.
- Release/close file.
- Make symlink.
- Read symlink.

## TODO

- Mount this filesystem with `mount` command.
- Copy on write, copy file from slaves to master backend before/while write.
- Copy on read(optional), copy file from slaves to master backend before/while write.
- Check operation permisson.
- Fix I/O error if files/dir removed from backend.
- etc.
