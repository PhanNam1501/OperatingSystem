package main

import (
	"os"

	"golang.org/x/sys/unix"
)

func main10() {
	args := os.Args
	err := unix.Unlink(args[1])
	if err != nil {
		panic(err)
	}
}
