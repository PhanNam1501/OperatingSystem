package main

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main6() {
	fd, err := unix.Open("test.txt", unix.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}
	fd2, err := unix.Dup(fd)
	if err != nil {
		panic(err)
	}
	buffer := make([]byte, 4096)
	r1, _, errNo := unix.RawSyscall(
		unix.SYS_READ,
		uintptr(fd2),
		uintptr(unsafe.Pointer(&buffer[0])),
		uintptr(len(buffer)),
	)
	if errNo != 0 {
		panic(errNo)
	}
	n := int(r1)
	fmt.Println(n)

}
