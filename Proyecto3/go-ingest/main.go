package main

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

type report struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
}

func validReport(item report) bool {
	switch item.Country {
	case "USA", "RUS", "CHN", "ESP", "GTM":
	default:
		return false
	}
	if item.WarplanesInAir < 0 || item.WarplanesInAir > 50 || item.WarshipsInWater < 0 || item.WarshipsInWater > 30 {
		return false
	}
	_, err := time.Parse(time.RFC3339Nano, item.Timestamp)
	return err == nil
}

func main() {
	clientURL := os.Getenv("CLIENT_URL")
	if clientURL == "" {
		clientURL = "http://127.0.0.1:8082/publish"
	}
	client := &http.Client{Timeout: 6 * time.Second}
	http.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		var input report
		if err := json.NewDecoder(io.LimitReader(r.Body, 1<<20)).Decode(&input); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if !validReport(input) {
			http.Error(w, "reporte incompleto", http.StatusBadRequest)
			return
		}
		body, _ := json.Marshal(input)
		request, err := http.NewRequestWithContext(r.Context(), http.MethodPost, clientURL, bytes.NewReader(body))
		if err != nil {
			http.Error(w, "cliente no disponible", http.StatusBadGateway)
			return
		}
		request.Header.Set("Content-Type", "application/json")
		response, err := client.Do(request)
		if err != nil {
			log.Printf("fallo cliente gRPC: %v", err)
			http.Error(w, "cliente no disponible", http.StatusBadGateway)
			return
		}
		defer response.Body.Close()
		if response.StatusCode != http.StatusAccepted {
			http.Error(w, "no se pudo procesar el reporte", http.StatusBadGateway)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("ok")) })
	server := &http.Server{Addr: ":8081", ReadHeaderTimeout: 5 * time.Second}
	log.Fatal(server.ListenAndServe())
}
