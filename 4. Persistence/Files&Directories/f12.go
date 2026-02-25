package main

import (
	"fmt"
	"log"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main12() {
	// Mở directory "."
	fd, err := unix.Open(".", unix.O_RDONLY|unix.O_DIRECTORY, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer unix.Close(fd)

	// buffer cho getdents
	buf := make([]byte, 4096)

	for {
		n, err := unix.Getdents(fd, buf)
		if err != nil {
			log.Fatal(err)
		}
		if n <= 0 {
			break
		}
		fmt.Println(buf)

		bpos := 0
		for bpos < n {
			dirent := (*unix.Dirent)(unsafe.Pointer(&buf[bpos]))

			nameBytes := buf[bpos+int(unsafe.Offsetof(dirent.Name)):]
			name := string(nameBytes[:clen(nameBytes)])

			if dirent.Ino != 0 {
				fmt.Printf("%d %s\n", dirent.Ino, name)
			}

			bpos += int(dirent.Reclen)
		}
	}
}

// tìm độ dài null-terminated string
func clen(b []byte) int {
	for i, v := range b {
		if v == 0 {
			return i
		}
	}
	return len(b)
}
