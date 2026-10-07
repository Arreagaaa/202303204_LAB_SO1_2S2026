package main

import (
	"bufio"
	"fmt"
	"net"
	"time"
)

type valkeyStore struct {
	address string
}

func newValkeyStore(address string) *valkeyStore {
	return &valkeyStore{address: address}
}

func (s *valkeyStore) save(item report) error {
	connection, err := net.DialTimeout("tcp", s.address, 2*time.Second)
	if err != nil {
		return err
	}
	defer connection.Close()
	key := "reports:" + item.Country
	value := fmt.Sprintf(`{"country":"%s","warplanes_in_air":%d,"warships_in_water":%d,"timestamp":"%s"}`,
		item.Country, item.WarplanesInAir, item.WarshipsInWater, item.Timestamp)
	if err := writeCommand(connection, "RPUSH", key, value); err != nil {
		return err
	}
	_, err = bufio.NewReader(connection).ReadString('\n')
	return err
}

func writeCommand(connection net.Conn, command string, args ...string) error {
	parts := append([]string{command}, args...)
	if _, err := fmt.Fprintf(connection, "*%d\r\n", len(parts)); err != nil {
		return err
	}
	for _, part := range parts {
		if _, err := fmt.Fprintf(connection, "$%d\r\n%s\r\n", len(part), part); err != nil {
			return err
		}
	}
	return nil
}
