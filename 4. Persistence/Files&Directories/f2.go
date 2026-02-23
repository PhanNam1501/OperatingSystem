package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main2() {
	fd1, _ := unix.Open("test.txt", unix.O_RDONLY, 0)
	fd2, _ := unix.Dup(fd1)
	fmt.Printf("FD1: %d, FD2: %d\n", fd1, fd2)

	unix.Close(fd1)
	unix.Close(fd2)
}
