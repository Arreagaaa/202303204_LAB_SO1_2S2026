package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
	pb "proyecto3/proto"
)

type writer struct {
	pb.UnimplementedWarReportServiceServer
	client *http.Client
	url    string
}

func (w *writer) SendReport(ctx context.Context, req *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	body, err := (protojson.MarshalOptions{UseProtoNames: true}).Marshal(req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Content-Type", "application/json")
	response, err := w.client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusAccepted {
		return nil, fmt.Errorf("publicador RabbitMQ respondio %d", response.StatusCode)
	}
	return &pb.WarReportResponse{Status: "published"}, nil
}

func main() {
	url := os.Getenv("PUBLISHER_URL")
	if url == "" {
		url = "http://127.0.0.1:8083/publish"
	}
	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal(err)
	}
	server := grpc.NewServer()
	pb.RegisterWarReportServiceServer(server, &writer{client: &http.Client{Timeout: 6 * time.Second}, url: url})
	log.Fatal(server.Serve(listener))
}
