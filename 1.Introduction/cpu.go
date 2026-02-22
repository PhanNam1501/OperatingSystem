package main

import (
	"fmt"
	"os"
	"time"
)

func spin(seconds int) {
	start := time.Now()
	for {
		if time.Since(start) >= time.Duration(seconds)*time.Second {
			return
		}
	}
}

func main1() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "usage: cpu <string>")
		os.Exit(1)
	}

	str := os.Args[1]
	for {
		spin(1)
		fmt.Println(str)
	}
}
