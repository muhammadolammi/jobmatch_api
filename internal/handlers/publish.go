package handlers

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func (config *Config) PublishSession(session Session) error {
	conn, err := amqp.Dial(config.RABBITMQUrl)
	if err != nil {
		return fmt.Errorf("error connecting to RabbitMQ. err:  %v", err)

	}
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("error getting connection channel. err: %v", err)

	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"sessions", // queue name
		true,       // durable
		false,      // auto-delete
		false,      // exclusive
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		return fmt.Errorf("error declating que on channel channel. err: %v", err)

	}

	body, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("error marshalling session. err: %v", err)
	}

	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key (queue name)
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("error publishing session. err: %v", err)

	}
	log.Println("session saved to db and published")
	return nil

}
