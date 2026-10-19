package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type publisher struct {
	mu       sync.Mutex
	channel  *amqp091.Channel
	confirms <-chan amqp091.Confirmation
	queue    string
}

func (p *publisher) publish(ctx context.Context, body []byte) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	sequence := p.channel.GetNextPublishSeqNo()
	if err := p.channel.PublishWithContext(ctx, "", p.queue, false, false, amqp091.Publishing{
		ContentType: "application/json", DeliveryMode: amqp091.Persistent, Body: body,
	}); err != nil {
		return err
	}
	for {
		select {
		case confirmation, ok := <-p.confirms:
			if !ok {
				return fmt.Errorf("se cerro el canal de confirmaciones")
			}
			if confirmation.DeliveryTag < sequence {
				continue
			}
			if confirmation.DeliveryTag != sequence || !confirmation.Ack {
				return fmt.Errorf("RabbitMQ no confirmo la publicacion")
			}
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}
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
		log.Fatal(err)
	}
	defer conn.Close()
	channel, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer channel.Close()
	if _, err := channel.QueueDeclare(queue, true, false, false, false, nil); err != nil {
		log.Fatal(err)
	}
	if err := channel.Confirm(false); err != nil {
		log.Fatal(err)
	}
	p := &publisher{channel: channel, confirms: channel.NotifyPublish(make(chan amqp091.Confirmation, 1)), queue: queue}
	http.HandleFunc("/publish", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "metodo no permitido", http.StatusMethodNotAllowed)
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
		if err != nil {
			http.Error(w, "reporte invalido", http.StatusBadRequest)
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()
		if err := p.publish(ctx, body); err != nil {
			log.Printf("fallo RabbitMQ: %v", err)
			http.Error(w, "RabbitMQ no disponible", http.StatusBadGateway)
			return
		}
		w.WriteHeader(http.StatusAccepted)
	})
	http.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	log.Fatal((&http.Server{Addr: ":8083", ReadHeaderTimeout: 5 * time.Second}).ListenAndServe())
}
