package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"os"

	"github.com/rabbitmq/amqp091-go"
	pb "proyecto3/proto"
	"google.golang.org/grpc"
)

type writer struct {
	pb.UnimplementedWarReportServiceServer
	channel *amqp091.Channel
	queue   string
}

func (w *writer) SendReport(ctx context.Context, req *pb.WarReportRequest) (*pb.WarReportResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	err = w.channel.PublishWithContext(ctx, "", w.queue, false, false, amqp091.Publishing{
		ContentType: "application/json", Body: body,
	})
	if err != nil {
		return nil, err
	}
	return &pb.WarReportResponse{Status: "published"}, nil
}

func main() {
	address := os.Getenv("RABBITMQ_ADDRESS")
	if address == "" {
		address = "amqp://guest:x@rabbitmq:5672/"
	}
	queue := os.Getenv("RABBITMQ_QUEUE")
	if queue == "" {
		queue = "military-reports"
	}
	conn, err := amqp091.Dial(address)
	if err != nil {
		log.Fatalf("no se pudo conectar con RabbitMQ: %v", err)
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("no se pudo abrir el canal: %v", err)
	}
	defer channel.Close()
	if _, err = channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatalf("no se pudo declarar la cola: %v", err)
	}

	listener, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("no se pudo abrir gRPC: %v", err)
	}
	server := grpc.NewServer(grpc.ForceServerCodec(jsonCodec{}))
	pb.RegisterWarReportServiceServer(server, &writer{channel: channel, queue: queue})
	log.Printf("go-writer escuchando en :50051")
	log.Fatal(server.Serve(listener))
}

type jsonCodec struct{}

func (jsonCodec) Name() string { return "json" }
func (jsonCodec) Marshal(value interface{}) ([]byte, error) { return json.Marshal(value) }
func (jsonCodec) Unmarshal(data []byte, value interface{}) error { return json.Unmarshal(data, value) }
