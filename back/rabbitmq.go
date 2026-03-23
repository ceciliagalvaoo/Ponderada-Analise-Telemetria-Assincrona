package main

import (
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitPublisher struct {
	channel   *amqp.Channel
	queueName string
}

func (r *RabbitPublisher) Publish(body []byte) error {
	return r.channel.Publish(
		"",
		r.queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent,
			Body:         body,
		},
	)
}

func connectRabbitMQ() (*amqp.Connection, *amqp.Channel, amqp.Queue) {
	var conn *amqp.Connection
	var err error

	for attempt := 1; attempt <= 120; attempt++ {
		conn, err = amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
		if err == nil {
			break
		}

		log.Printf("tentativa %d/120: erro ao conectar no RabbitMQ: %v", attempt, err)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatal("erro ao conectar no RabbitMQ apos tentativas: ", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("erro ao abrir canal: ", err)
	}

	q, err := ch.QueueDeclare(
		"telemetry_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("erro ao declarar fila: ", err)
	}

	return conn, ch, q
}
