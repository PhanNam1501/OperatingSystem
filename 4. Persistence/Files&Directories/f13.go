package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

func statFile(path string) {
	var stat unix.Stat_t
	err := unix.Stat(path, &stat)
	if err != nil {
		log.Printf("stat %s error: %v\n", path, err)
		return
	}
	fmt.Printf("File: %s\n", path)
	fmt.Printf("  Inode: %d\n", stat.Ino)
	fmt.Printf("  Links: %d\n", stat.Nlink)
	fmt.Println()
}

func main13() {
	fd, err := unix.Open("test.txt", unix.O_CREAT|unix.O_RDWR, 0644)
	if err != nil {
		log.Fatal(err)
	}
	unix.Write(fd, []byte("hello\n"))
	unix.Close(fd)

	fmt.Println("After creating file:")
	statFile("test.txt")

	err = unix.Link("test.txt", "test1.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("After creating hard link test1:")
	statFile("test.txt")
	statFile("test1.txt")

	err = unix.Link("test1.txt", "test2.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("After creating hard link test2:")
	statFile("test.txt")
	statFile("test1.txt")
	statFile("test2.txt")

	err = unix.Unlink("test2.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("After unlink file:")
	statFile("test.txt")
	statFile("test1.txt")

	err = unix.Unlink("test1.txt")
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("After unlink test1:")
	statFile("test.txt")

}
