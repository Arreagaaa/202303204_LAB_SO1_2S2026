package valkey

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	conn net.Conn
	rw   *bufio.Reader
}

func Dial(address string) (*Client, error) {
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return nil, err
	}
	return &Client{conn: conn, rw: bufio.NewReader(conn)}, nil
}

func (client *Client) Close() error {
	if client == nil || client.conn == nil {
		return nil
	}
	return client.conn.Close()
}

func (client *Client) Set(key, value string) error {
	_, err := client.do("SET", key, value)
	return err
}

func (client *Client) RPush(key, value string) error {
	_, err := client.do("RPUSH", key, value)
	return err
}

func (client *Client) do(command string, args ...string) (string, error) {
	if client == nil || client.conn == nil {
		return "", fmt.Errorf("cliente valkey no inicializado")
	}
	if err := writeCommand(client.conn, command, args...); err != nil {
		return "", err
	}
	return readReply(client.rw)
}

func (client *Client) SetInteger(key string, value uint64) error {
	_, err := client.do("SET", key, fmt.Sprintf("%d", value))
	return err
}

func writeCommand(conn net.Conn, command string, args ...string) error {
	parts := append([]string{command}, args...)
	if _, err := fmt.Fprintf(conn, "*%d\r\n", len(parts)); err != nil {
		return err
	}
	for _, part := range parts {
		if _, err := fmt.Fprintf(conn, "$%d\r\n%s\r\n", len(part), part); err != nil {
			return err
		}
	}
	return nil
}

func readReply(rw *bufio.Reader) (string, error) {
	line, err := rw.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
	if len(line) == 0 {
		return "", fmt.Errorf("respuesta vacía")
	}
	switch line[0] {
	case '+':
		return line[1:], nil
	case ':':
		return line[1:], nil
	case '-':
		return "", fmt.Errorf(strings.TrimSpace(line[1:]))
	case '$':
		length, err := strconv.Atoi(line[1:])
		if err != nil {
			return "", err
		}
		if length < 0 {
			return "", nil
		}
		payload := make([]byte, length+2)
		if _, err := io.ReadFull(rw, payload); err != nil {
			return "", err
		}
		return string(payload[:length]), nil
	default:
		return "", fmt.Errorf("respuesta RESP desconocida: %q", line)
	}
}
