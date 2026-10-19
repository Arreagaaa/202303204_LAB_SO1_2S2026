package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	pb "proyecto3/proto"
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
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewWarReportServiceClient(conn)
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		var input report
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "JSON invalido", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		_, err := client.SendReport(ctx, &pb.WarReportRequest{
			Country: input.Country, WarplanesInAir: input.WarplanesInAir,
			WarshipsInWater: input.WarshipsInWater, Timestamp: input.Timestamp,
		})
		if err != nil {
			log.Printf("fallo gRPC: %v", err)
			http.Error(w, "writer no disponible", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	log.Fatal((&http.Server{Addr: ":8082", ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
}
