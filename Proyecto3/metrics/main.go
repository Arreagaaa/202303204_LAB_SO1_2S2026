package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type report struct {
	WarplanesInAir  int `json:"warplanes_in_air"`
	WarshipsInWater int `json:"warships_in_water"`
}

var countries = []string{"USA", "RUS", "CHN", "ESP", "GTM"}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadString('\n')
	return strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r"), err
}

func readReports(reader *bufio.Reader) ([]report, error) {
	header, err := readLine(reader)
	if err != nil || !strings.HasPrefix(header, "*") {
		return nil, fmt.Errorf("respuesta RESP invalida: %q: %w", header, err)
	}
	count, err := strconv.Atoi(header[1:])
	if err != nil || count < 0 {
		return nil, fmt.Errorf("cantidad RESP invalida: %q", header)
	}
	reports := make([]report, 0, count)
	for i := 0; i < count; i++ {
		lengthLine, err := readLine(reader)
		if err != nil || !strings.HasPrefix(lengthLine, "$") {
			return nil, fmt.Errorf("longitud RESP invalida: %q: %w", lengthLine, err)
		}
		length, err := strconv.Atoi(lengthLine[1:])
		if err != nil || length < 0 || length > 1<<20 {
			return nil, fmt.Errorf("longitud de reporte invalida: %q", lengthLine)
		}
		body := make([]byte, length+2)
		if _, err := io.ReadFull(reader, body); err != nil {
			return nil, err
		}
		var item report
		if err := json.Unmarshal(body[:length], &item); err != nil {
			return nil, err
		}
		reports = append(reports, item)
	}
	return reports, nil
}

func fetch(address, country string) ([]report, error) {
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return nil, err
	}
	defer conn.Close()
	if err := conn.SetDeadline(time.Now().Add(10 * time.Second)); err != nil {
		return nil, err
	}
	key := "reports:" + country
	if _, err := fmt.Fprintf(conn, "*4\r\n$6\r\nLRANGE\r\n$%d\r\n%s\r\n$1\r\n0\r\n$2\r\n-1\r\n", len(key), key); err != nil {
		return nil, err
	}
	return readReports(bufio.NewReader(conn))
}

func mode(counts map[int]int) int {
	values := make([]int, 0, len(counts))
	for value := range counts {
		values = append(values, value)
	}
	sort.Ints(values)
	best, frequency := 0, 0
	for _, value := range values {
		if counts[value] > frequency {
			best, frequency = value, counts[value]
		}
	}
	return best
}

func metrics(w http.ResponseWriter, _ *http.Request) {
	address := os.Getenv("VALKEY_ADDRESS")
	if address == "" {
		address = "valkey-reports:6379"
	}
	all := make([]report, 0)
	assigned := []report{}
	type totals struct{ reports, planes, ships int }
	byCountry := make(map[string]totals)
	for _, country := range countries {
		items, err := fetch(address, country)
		if err != nil {
			http.Error(w, err.Error(), http.StatusServiceUnavailable)
			return
		}
		planes, ships := 0, 0
		for _, item := range items {
			planes += item.WarplanesInAir
			ships += item.WarshipsInWater
		}
		byCountry[country] = totals{len(items), planes, ships}
		all = append(all, items...)
		if country == "CHN" {
			assigned = items
		}
	}
	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	for _, country := range countries {
		total := byCountry[country]
		fmt.Fprintf(w, "military_country_reports_total{country=%q} %d\n", country, total.reports)
		fmt.Fprintf(w, "military_country_warplanes_total{country=%q} %d\n", country, total.planes)
		fmt.Fprintf(w, "military_country_warships_total{country=%q} %d\n", country, total.ships)
	}
	if len(all) == 0 {
		fmt.Fprintln(w, "military_reports_total 0")
		return
	}
	minPlanes, maxPlanes := all[0].WarplanesInAir, all[0].WarplanesInAir
	minShips, maxShips := all[0].WarshipsInWater, all[0].WarshipsInWater
	planeCounts, shipCounts := map[int]int{}, map[int]int{}
	for _, item := range all {
		if item.WarplanesInAir < minPlanes {
			minPlanes = item.WarplanesInAir
		}
		if item.WarplanesInAir > maxPlanes {
			maxPlanes = item.WarplanesInAir
		}
		if item.WarshipsInWater < minShips {
			minShips = item.WarshipsInWater
		}
		if item.WarshipsInWater > maxShips {
			maxShips = item.WarshipsInWater
		}
		planeCounts[item.WarplanesInAir]++
		shipCounts[item.WarshipsInWater]++
	}
	fmt.Fprintf(w, "military_reports_total %d\n", len(all))
	fmt.Fprintf(w, "military_warplanes_max %d\n", maxPlanes)
	fmt.Fprintf(w, "military_warplanes_min %d\n", minPlanes)
	fmt.Fprintf(w, "military_warships_max %d\n", maxShips)
	fmt.Fprintf(w, "military_warships_min %d\n", minShips)
	fmt.Fprintf(w, "military_warplanes_mode %d\n", mode(planeCounts))
	fmt.Fprintf(w, "military_warships_mode %d\n", mode(shipCounts))
	if len(assigned) > 0 {
		latest := assigned[len(assigned)-1]
		fmt.Fprintf(w, "military_assigned_warplanes %d\n", latest.WarplanesInAir)
		fmt.Fprintf(w, "military_assigned_warships %d\n", latest.WarshipsInWater)
	}
}

func main() {
	http.HandleFunc("/metrics", metrics)
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { fmt.Fprint(w, "ok") })
	log.Fatal(http.ListenAndServe(":9091", nil))
}
