package main

import (
	"fmt"
	"net"
	"testing"
)

func TestSaveRejectsValkeyError(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		defer conn.Close()
		_, _ = fmt.Fprint(conn, "-ERR write failed\r\n")
	}()
	store := newValkeyStore(listener.Addr().String())
	if err := store.save(report{Country: "CHN", Timestamp: "2026-10-19T00:00:00Z"}); err == nil {
		t.Fatal("save accepted a Valkey error response")
	}
}
