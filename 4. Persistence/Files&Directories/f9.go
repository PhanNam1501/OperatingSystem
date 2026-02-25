package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func atomicWrite(filename string, data []byte) error {
	tmpName := filename + ".tmp"
	fd, err := unix.Open(tmpName,
		unix.O_WRONLY|unix.O_CREAT|unix.O_TRUNC,
		0600,
	)
	if err != nil {
		return err
	}

	if _, err := unix.Write(fd, data); err != nil {
		unix.Close(fd)
		return err
	}

	if err := unix.Fsync(fd); err != nil {
		unix.Close(fd)
		return err
	}

	if err := unix.Close(fd); err != nil {
		return err
	}

	if err := unix.Rename(tmpName, filename); err != nil {
		return err
	}

	dirfd, err := unix.Open(".", unix.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer unix.Close(dirfd)
	if err := unix.Fsync(dirfd); err != nil {
		return err
	}

	return nil
}

func main9() {
	args := os.Args
	data := []byte("Hello everyone.")
	atomicWrite(args[1], data)
}
