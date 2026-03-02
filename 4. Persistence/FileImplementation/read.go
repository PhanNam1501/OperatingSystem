package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

func test2() {
	fd, err := unix.Open("test.txt", unix.O_RDONLY, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer unix.Close(fd)

	buf := make([]byte, 10)

	// Đọc 10 byte từ offset 100
	n, err := unix.Pread(fd, buf, 100)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Đọc:", n, "bytes")
	fmt.Println(string(buf[:n]))
}

func test1() {
	fd, err := unix.Open("test.txt", unix.O_RDONLY, 0)
	if err != nil {
		log.Fatal(err)
	}
	defer unix.Close(fd)

	buf := make([]byte, 4096)

	for {
		n, err := unix.Read(fd, buf)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			if n == 0 {
				break
			}
			log.Fatal(err)
		}

		if n == 0 {
			break
		}

		fmt.Printf("Đọc %d bytes\n", n)
		fmt.Println(string(buf[:n]))
	}
}

func main2() {
	test1()
	test2()
}
