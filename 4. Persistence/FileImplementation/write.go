package main

import (
	"fmt"
	"log"

	"golang.org/x/sys/unix"
)

func main() {
	fd, err := unix.Open("test.txt",
		unix.O_CREAT|unix.O_WRONLY|unix.O_TRUNC,
		0644)
	if err != nil {
		log.Fatal(err)
	}
	defer unix.Close(fd)

	data := []byte("Safe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write exampleSafe write example\n")

	if err := writeAll(fd, data); err != nil {
		log.Fatal(err)
	}
}

func writeAll(fd int, data []byte) error {
	total := 0
	count := 0
	for total < len(data) {
		n, err := unix.Write(fd, data[total:])
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return err
		}
		total += n
		count++
	}
	fmt.Println(count)
	return nil
}
