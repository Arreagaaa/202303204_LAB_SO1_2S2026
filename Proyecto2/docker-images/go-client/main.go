package main

import (
	"os"
	"runtime"
	"time"
)

func main() {
	blocks := make([][]byte, 0, 32)
	pageSize := os.Getpagesize()
	for len(blocks) < 32 {
		block := make([]byte, 16*1024*1024)
		for page := 0; page < len(block); page += pageSize {
			block[page] = 1
		}
		blocks = append(blocks, block)
		runtime.GC()
		time.Sleep(250 * time.Millisecond)
	}
	for {
		time.Sleep(time.Hour)
	}
}
