package main

import (
	"encoding/json"
	"log"
	"os"
	"time"

	"github.com/rabbitmq/amqp091-go"
)

type report struct {
	Country         string `json:"country"`
	WarplanesInAir  int32  `json:"warplanes_in_air"`
	WarshipsInWater int32  `json:"warships_in_water"`
	Timestamp       string `json:"timestamp"`
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
	var conn *amqp091.Connection
	var err error
	for {
		conn, err = amqp091.Dial(address)
		if err == nil {
			break
		}
		log.Printf("RabbitMQ no disponible; reintentando en 2 segundos: %v", err)
		time.Sleep(2 * time.Second)
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
	messages, err := channel.Consume(queue, "mumnk8s-consumer", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("no se pudo consumir la cola: %v", err)
	}

	valkeyAddress := os.Getenv("VALKEY_ADDRESS")
	if valkeyAddress == "" {
		valkeyAddress = "valkey:6379"
	}
	store := newValkeyStore(valkeyAddress)
	log.Printf("consumer esperando mensajes")
	for message := range messages {
		var item report
		if err := json.Unmarshal(message.Body, &item); err != nil {
			log.Printf("mensaje invalido: %v", err)
			_ = message.Nack(false, false)
			continue
		}
		if err := store.save(item); err != nil {
			log.Printf("no se pudo guardar reporte: %v", err)
			_ = message.Nack(false, true)
			time.Sleep(time.Second)
			continue
		}
		_ = message.Ack(false)
	}
}
