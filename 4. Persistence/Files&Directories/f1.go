package main

import (
	"fmt"
	"os"
)

func main1() {
	file, err := os.OpenFile("test.txt", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		panic(err)
	}

	defer file.Close()

	file.Write([]byte("hello\n"))

	fd := file.Fd()
	fmt.Printf("file descriptor: %d\n", fd)
}
