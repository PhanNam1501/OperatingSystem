package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func main11() {
	args := os.Args
	err := unix.Rmdir(args[1])
	if err != nil {
		panic(err)
	}

}
