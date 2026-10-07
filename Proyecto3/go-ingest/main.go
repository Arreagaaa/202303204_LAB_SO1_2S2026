package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	pb "proyecto3/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type report struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
}

func main() {
	target := os.Getenv("WRITER_ADDRESS")
	if target == "" {
		target = "go-writer:50051"
	}

	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("no se pudo conectar con el escritor: %v", err)
	}
	defer conn.Close()

	client := pb.NewWarReportServiceClient(conn)
	http.HandleFunc("/reports", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		var input report
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		if input.Country == "" || input.Timestamp == "" || input.WarplanesInAir < 0 || input.WarshipsInWater < 0 {
			http.Error(w, "reporte incompleto", http.StatusBadRequest)
			return
		}

		_, err := client.SendReport(context.Background(), &pb.WarReportRequest{
			Country: input.Country, WarplanesInAir: input.WarplanesInAir,
			WarshipsInWater: input.WarshipsInWater, Timestamp: input.Timestamp,
		})
		if err != nil {
			log.Printf("fallo al enviar reporte: %v", err)
			http.Error(w, "no se pudo procesar el reporte", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	server := &http.Server{Addr: ":8081", ReadHeaderTimeout: 5 * time.Second}
	log.Printf("go-ingest escuchando en %s", server.Addr)
	log.Fatal(server.ListenAndServe())
}
