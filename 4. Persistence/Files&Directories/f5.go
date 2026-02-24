package main

import (
	"fmt"
	"os"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main5() {
	fd, err := unix.Open("test.txt", unix.O_RDONLY, 0)
	if err != nil {
		panic(err)
	}

	pid, _, errNo := unix.RawSyscall(unix.SYS_FORK, 0, 0, 0)
	if errNo != 0 {
		panic(errNo)
	}

	if pid == 0 {
		offset, err := unix.Seek(fd, 0, unix.SEEK_SET)
		if err != nil {
			panic(err)
		}
		fmt.Printf("child: offset %d\n", offset)
		buffer := make([]byte, 4096)
		r1, _, errNo := unix.RawSyscall(
			unix.SYS_READ,
			uintptr(fd),
			uintptr(unsafe.Pointer(&buffer[0])),
			uintptr(len(buffer)),
		)
		n := int(r1)
		fmt.Printf("Read %d byte: %s\n", n, string(buffer[:n]))

		if errNo != 0 {
			panic(errNo)
		}
		os.Exit(0)
	} else {
		_, err := unix.Wait4(int(pid), nil, 0, nil)
		if err != nil {
			panic(err)
		}

		offset, err := unix.Seek(fd, 10, unix.SEEK_CUR)
		if err != nil {
			panic(err)
		}
		fmt.Printf("parent: offset %d\n", offset)
	}
}
