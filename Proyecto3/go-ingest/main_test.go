package main

import "testing"

func TestValidReport(t *testing.T) {
	good := report{Country: "CHN", WarplanesInAir: 50, WarshipsInWater: 30, Timestamp: "2026-10-19T12:00:00Z"}
	if !validReport(good) {
		t.Fatal("valid report rejected")
	}
	bad := good
	bad.Country = "MEX"
	if validReport(bad) {
		t.Fatal("unknown country accepted")
	}
	bad = good
	bad.WarshipsInWater = 31
	if validReport(bad) {
		t.Fatal("out-of-range ships accepted")
	}
}
