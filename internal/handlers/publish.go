package handlers

import (
	"encoding/json"
	"fmt"

	"github.com/streadway/amqp"
)

func (apiConfig *Config) PublishSession(session Session, rabbitConn *amqp.Connection) error {

	ch, err := rabbitConn.Channel()
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
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return fmt.Errorf("error publishing session. err: %v", err)

	}
	// log.Println("session saved to db and published")
	return nil

}
