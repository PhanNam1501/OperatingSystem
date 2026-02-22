package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// Tương đương malloc(sizeof(int))
	p := new(int) // cấp phát trên heap

	// Go không cần assert vì new() luôn trả về con trỏ hợp lệ
	fmt.Printf("(%d) address pointed to by p: %p\n",
		os.Getpid(), p)

	*p = 0

	for {
		time.Sleep(1 * time.Second) // tương đương Spin(1)
		*p = *p + 1
		fmt.Printf("(%d) p: %d\n", os.Getpid(), *p)
	}
}
