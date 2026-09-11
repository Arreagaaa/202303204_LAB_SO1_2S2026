package main

import (
	"runtime"
	"time"
)

func main() {
	blocks := make([][]byte, 0, 32)
	for len(blocks) < 32 {
		blocks = append(blocks, make([]byte, 16*1024*1024))
		runtime.GC()
		time.Sleep(250 * time.Millisecond)
	}
	select {}
}
