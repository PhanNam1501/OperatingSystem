package main

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"unsafe"

	"golang.org/x/sys/unix"
)

func main8() {
	if len(os.Args) < 2 {
		return
	}

	// Open file
	fd, err := unix.Open("ps.img", unix.O_RDWR, 0)
	if err != nil {
		panic(err)
	}
	defer unix.Close(fd)

	// Stat file
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		panic(err)
	}
	fileSize := int(stat.Size)

	// Mmap
	data, err := unix.Mmap(fd, 0, fileSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		panic(err)
	}
	defer unix.Munmap(data)

	// First 8 bytes = n (uint64)
	nPtr := (*uint64)(unsafe.Pointer(&data[0]))
	n := *nPtr
	fmt.Println(n)

	// Remaining bytes = stack of int32
	stackOffset := int(unsafe.Sizeof(uint64(0)))
	stackCapacity := (fileSize - stackOffset) / 4
	fmt.Println(stackOffset)
	fmt.Println(stackCapacity)

	for i := 1; i < len(os.Args); i++ {

		if os.Args[i] == "pop" {

			if n > 0 {
				n--
				offset := stackOffset + int(n)*4
				val := int32(binary.LittleEndian.Uint32(data[offset:]))
				fmt.Println(val)
				*nPtr = n
			}

		} else {

			if int(n) < stackCapacity {
				val, err := strconv.Atoi(os.Args[i])
				if err != nil {
					continue
				}

				offset := stackOffset + int(n)*4
				binary.LittleEndian.PutUint32(data[offset:], uint32(int32(val)))
				n++
				*nPtr = n
			}
		}
	}
}
