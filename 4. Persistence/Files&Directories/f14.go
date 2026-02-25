package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

func main() {
	fd, _ := unix.Open("test.txt", unix.O_CREAT|unix.O_RDWR, 0644)
	unix.Write(fd, []byte("ETH TOC"))
	var s unix.Stat_t
	unix.Stat("test.txt", &s)
	fmt.Println("Inode:", s.Ino)
	fmt.Println("Mode:", s.Mode)
	unix.Close(fd)

	err := unix.Symlink("test.txt", "test1.txt")
	if err != nil {
		log.Fatal(err)
	}

	var stat unix.Stat_t
	unix.Lstat("test1.txt", &stat)

	fmt.Println("Inode:", stat.Ino)
	fmt.Println("Mode:", stat.Mode)

	err = unix.Unlink("test1.txt")
	// err = unix.Unlink("test2.txt")
	// err = unix.Unlink("test3.txt")

}
