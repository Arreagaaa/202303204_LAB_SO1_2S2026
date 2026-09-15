package valkey

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"testing"
)

func TestSetAndRPushWriteRESPCommands(t *testing.T) {
	t.Parallel()

	server, clientConn := net.Pipe()
	defer server.Close()
	defer clientConn.Close()

	client := &Client{conn: clientConn, rw: bufio.NewReader(clientConn)}
	responses := make(chan error, 1)
	go func() {
		reader := bufio.NewReader(server)
		if command, err := readRESPCommand(reader); err != nil {
			responses <- err
			return
		} else if command != "SET system:ram:current total=1KB" {
			responses <- fmt.Errorf("unexpected first command: %q", command)
			return
		}
		fmt.Fprint(server, "+OK\r\n")
		if command, err := readRESPCommand(reader); err != nil {
			responses <- err
			return
		} else if command != "RPUSH system:ram:history 1|0|1" {
			responses <- fmt.Errorf("unexpected second command: %q", command)
			return
		}
		fmt.Fprint(server, "+OK\r\n")
		responses <- nil
	}()

	if err := client.Set("system:ram:current", "total=1KB"); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := client.RPush("system:ram:history", "1|0|1"); err != nil {
		t.Fatalf("RPush() error = %v", err)
	}
	if err := server.Close(); err != nil {
		t.Fatalf("close server: %v", err)
	}
	if err := <-responses; err != nil {
		t.Fatalf("server read error: %v", err)
	}
}

func readRESPCommand(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	if len(line) == 0 || line[0] != '*' {
		return "", fmt.Errorf("expected array header, got %q", line)
	}
	parts := make([]string, 0, 3)
	for index := 0; index < 3; index++ {
		head, err := reader.ReadString('\n')
		if err != nil {
			return "", err
		}
		if len(head) == 0 || head[0] != '$' {
			return "", fmt.Errorf("expected bulk header, got %q", head)
		}
		length, err := strconv.Atoi(strings.TrimSuffix(strings.TrimSuffix(head[1:], "\n"), "\r"))
		if err != nil {
			return "", err
		}
		payload := make([]byte, length+2)
		if _, err := io.ReadFull(reader, payload); err != nil {
			return "", err
		}
		parts = append(parts, string(payload[:length]))
	}
	return fmt.Sprintf("%s %s %s", parts[0], parts[1], parts[2]), nil
}

func TestReadReplyHandlesSimpleStrings(t *testing.T) {
	t.Parallel()

	server, clientConn := net.Pipe()
	defer server.Close()
	defer clientConn.Close()

	go func() {
		_, _ = fmt.Fprint(server, "+OK\r\n")
	}()

	reply, err := readReply(bufio.NewReader(clientConn))
	if err != nil {
		t.Fatalf("readReply() error = %v", err)
	}
	if reply != "OK" {
		t.Fatalf("readReply() = %q, want OK", reply)
	}
}
