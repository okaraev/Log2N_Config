package main

import (
	"encoding/json"

	"github.com/streadway/amqp"
)

func SendMessage(message interface{}, configParams interface{}) error {
	confParams := configParams.(commonconfig)
	connectRabbitMQ, err := amqp.Dial(confParams.QConnectionString)
	if err != nil {
		return err
	}
	defer connectRabbitMQ.Close()
	channelRabbitMQ, err := connectRabbitMQ.Channel()
	if err != nil {
		return err
	}
	defer channelRabbitMQ.Close()
	bytes, err := json.Marshal(message)
	if err != nil {
		return err
	}
	mess := amqp.Publishing{
		ContentType:  "application/json",
		Body:         bytes,
		DeliveryMode: 2,
	}
	err = channelRabbitMQ.Publish("", confParams.QName, false, false, mess)
	if err != nil {
		return err
	}
	return nil
}
