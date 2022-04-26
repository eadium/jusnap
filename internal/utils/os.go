package utils

import (
	"fmt"
	"io"
	"os"
)

func Copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}

func SetFileMod(fname string, perms os.FileMode, uid, gid int) error {
	err := os.Chmod(fname, perms)
	if err != nil {
		return nil
	}
	if gid == 0 || uid == 0 {
		uid, gid = os.Getuid(), os.Getgid()
	}
	err = os.Chown(fname, uid, gid)
	if err != nil {
		return nil
	}
	return nil
}
