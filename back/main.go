package main

import (
	"log"
	"net/http"
)

func main() {
	conn, ch, q := connectRabbitMQ()
	defer conn.Close()
	defer ch.Close()

	publisher := &RabbitPublisher{
		channel:   ch,
		queueName: q.Name,
	}

	http.HandleFunc("/telemetry", telemetryHandler(publisher))

	log.Println("back rodando na porta 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}