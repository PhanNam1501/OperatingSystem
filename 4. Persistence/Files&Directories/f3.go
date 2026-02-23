package main

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

func main3() {
	entries, err := os.ReadDir(".")
	if err != nil {
		panic(err)
	}

	for _, entry := range entries {
		var stat unix.Stat_t

		err := unix.Stat(entry.Name(), &stat)
		if err != nil {
			continue
		}
		fmt.Printf("Inode: %d | Name: %s\n", stat.Ino, entry.Name())
	}
}
