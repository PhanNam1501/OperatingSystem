package main

import (
	"fmt"

	"golang.org/x/sys/unix"
)

func main7() {
	fd, err := unix.Open(
		"test.txt",
		unix.O_CREAT|unix.O_WRONLY|unix.O_TRUNC,
		0600,
	)
	if err != nil {
		panic(err)
	}
	data := []byte("Hello Minh Ngoc Beo")
	rc, err := unix.Write(fd, data)
	if err != nil {
		panic(err)
	}
	fmt.Println(rc)

}
