package handlers

import (
	"context"
	"encoding/json"
	"log"
	"sync"

	_ "github.com/lib/pq"
	"github.com/muhammadolammi/jobmatchapi/internal/agent"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/streadway/amqp"
)

func callAgent(session Session) error {
	// TODO COMPLETE THE CALL AGENT USING GOOGLE ADK
	analyzer, err := agent.GetAgent()
	if err != nil {

	}
	return nil
}
func worker(id int, config *Config, wg *sync.WaitGroup) {
	defer wg.Done()
	//    to consume message on the queue
	conn, err := amqp.Dial(config.RABBITMQUrl)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	msgs, err := ch.Consume(
		"sessions", // queue name
		"",         // consumer tag
		true,       // auto-ack
		false,      // exclusive
		false,      // no-local
		false,      // no-wait
		nil,        // arguments
	)
	if err != nil {
		log.Fatal(err)
	}

	for msg := range msgs {
		// Here you can call a function to send the actual session
		// Unmarshal the body
		session := Session{}
		err = json.Unmarshal(msg.Body, &session)
		if err != nil {
			log.Printf("error unmarshalling message body. err: %v", err)
			continue
		}
		log.Printf("Worker %d processing session. session_id: %s", id+1, session.ID)

		err = callAgent(session)
		if err != nil {
			log.Printf("error running agent for session_id: %v. err: %v", session.ID, err)
			continue
		}
		// update session status
		err = config.DB.UpdateSessionStatus(context.Background(), database.UpdateSessionStatusParams{
			Status: "completed",
			ID:     session.ID,
		})
		if err != nil {
			log.Printf("error updating session status in db to completed for  session_id: %v. err: %v", session.ID, err)
			continue
		}
	}

}

func (config *Config) StartConsumerWorkerPool(numWorkers int) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := range numWorkers {
		log.Println("worker id ", i+1, "started")
		// wg.Done()
		// continue
		go worker(i, config, &wg)
	}
	wg.Wait() // block until all workers finish

}
