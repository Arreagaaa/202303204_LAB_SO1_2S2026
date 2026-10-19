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
	if err := connection.SetDeadline(time.Now().Add(5 * time.Second)); err != nil {
		return err
	}
	key := "reports:" + item.Country
	value := fmt.Sprintf(`{"country":"%s","warplanes_in_air":%d,"warships_in_water":%d,"timestamp":"%s"}`,
		item.Country, item.WarplanesInAir, item.WarshipsInWater, item.Timestamp)
	if err := writeCommand(connection, "RPUSH", key, value); err != nil {
		return err
	}
	response, err := bufio.NewReader(connection).ReadString('\n')
	if err != nil {
		return err
	}
	if len(response) == 0 || response[0] != ':' {
		return fmt.Errorf("respuesta inesperada de Valkey: %s", response)
	}
	return nil
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
