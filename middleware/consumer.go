package main

import (
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

func consumeMessages(repo DBInserter) {
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	if err != nil {
		log.Fatal("erro ao conectar no RabbitMQ:", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal("erro ao abrir canal:", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"telemetry_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("erro ao declarar fila:", err)
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("erro ao consumir fila:", err)
	}

	log.Println("middleware consumindo mensagens...")

	for msg := range msgs {
		if err := processMessage(repo, msg.Body); err != nil {
			log.Println("erro ao processar mensagem:", err)
			if nackErr := msg.Nack(false, true); nackErr != nil {
				log.Println("erro ao reenfileirar mensagem:", nackErr)
			}
			continue
		}

		if ackErr := msg.Ack(false); ackErr != nil {
			log.Println("erro ao confirmar mensagem:", ackErr)
			continue
		}

		log.Println("telemetria salva com sucesso")
	}
}
