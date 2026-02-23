package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main4() {
	fd, err := unix.Open(
		"example.txt",
		unix.O_CREAT|unix.O_RDWR|unix.O_TRUNC,
		0644,
	)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	data := []byte("Hello OS World")
	n, err := unix.Write(fd, data)
	if err != nil {
		panic(err)
	}
	fmt.Println("Bytes written:", n)

	offset, err := unix.Seek(fd, 0, unix.SEEK_SET)
	if err != nil {
		panic(err)
	}
	fmt.Println("New offset:", offset)

	buf := make([]byte, 100)
	n, err = unix.Read(fd, buf)
	if err != nil {
		panic(err)
	}

	fmt.Println("Bytes read:", n)
	fmt.Println("Content:", string(buf[:n]))
}
