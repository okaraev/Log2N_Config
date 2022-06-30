package main

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

func SendMessage(amqpServerURL string, QName string, i interface{}) error {
	connectRabbitMQ, err := amqp.Dial(amqpServerURL)
	if err != nil {
		return err
	}
	defer connectRabbitMQ.Close()
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channelRabbitMQ.Close()
	bytes, err := json.Marshal(i)
	if err != nil {
		return err
	}
	message := amqp.Publishing{
		ContentType:  "application/json",
		Body:         bytes,
		DeliveryMode: 2,
	}
	err = channelRabbitMQ.Publish("", QName, false, false, message)
	if err != nil {
		return err
	}
	return nil
}
