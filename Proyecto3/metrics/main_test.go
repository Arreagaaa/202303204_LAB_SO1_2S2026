package main

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestMetricsFromValkey(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()
	go func() {
		for i := 0; i < len(countries); i++ {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			reader := bufio.NewReader(conn)
			_, _ = reader.ReadString('\n')
			for part := 0; part < 4; part++ {
				line, err := reader.ReadString('\n')
				if err != nil {
					break
				}
				length, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "$")))
				if err != nil {
					break
				}
				_, _ = io.CopyN(io.Discard, reader, int64(length+2))
			}
			body := `{"warplanes_in_air":42,"warships_in_water":14}`
			_, _ = fmt.Fprintf(conn, "*1\r\n$%d\r\n%s\r\n", len(body), body)
			_ = conn.Close()
		}
	}()
	t.Setenv("VALKEY_ADDRESS", listener.Addr().String())
	response := httptest.NewRecorder()
	metrics(response, httptest.NewRequest("GET", "/metrics", nil))
	if response.Code != 200 {
		t.Fatalf("unexpected response: %d %s", response.Code, response.Body.String())
	}
	for _, expected := range []string{
		"military_reports_total 5",
		"military_warplanes_max 42",
		"military_warships_mode 14",
		"military_assigned_warplanes 42",
		`military_country_reports_total{country="CHN"} 1`,
	} {
		if !strings.Contains(response.Body.String(), expected) {
			t.Errorf("missing %q in metrics output", expected)
		}
	}
}
