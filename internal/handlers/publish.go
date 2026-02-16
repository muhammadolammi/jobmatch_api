package handlers

import (
	"context"
	"log"

	"cloud.google.com/go/pubsub/v2"
)

// func (apiConfig *Config) PublishSession(session Session, rabbitChan *amqp.Channel) error {
func (apiConfig *Config) PublishSession(session Session) error {

	// defer rabbitChan.Close()

	// q, err := rabbitChan.QueueDeclare(
	// 	"sessions", // queue name
	// 	true,       // durable
	// 	false,      // auto-delete
	// 	false,      // exclusive
	// 	false,      // no-wait
	// 	nil,        // arguments
	// )
	// if err != nil {
	// 	return fmt.Errorf("error declating que on channel channel. err: %v", err)

	// }

	// body, err := json.Marshal(session)
	// if err != nil {
	// 	return fmt.Errorf("error marshalling session. err: %v", err)
	// }

	// err = rabbitChan.Publish(
	// 	"",     // exchange
	// 	q.Name, // routing key (queue name)
	// 	false,  // mandatory
	// 	false,  // immediate
	// 	amqp.Publishing{
	// 		ContentType: "application/json",
	// 		Body:        body,
	// 	})
	// if err != nil {
	// 	return fmt.Errorf("error publishing session. err: %v", err)

	// }
	// // log.Println("session saved to db and published")
	// return nil

	// lets send a simple http post request to the worker instead of using rabbitmq for now, we can switch to rabbitmq later if needed
	// workerURL := apiConfig.WorkerApi + "/process-session"
	// jsonData, err := json.Marshal(session)
	// if err != nil {
	// 	return fmt.Errorf("error marshalling session. err: %v", err)
	// }
	// req, err := http.NewRequest("POST", workerURL, bytes.NewBuffer(jsonData))
	// if err != nil {
	// 	return fmt.Errorf("error creating request to worker. err: %v", err)
	// }
	// req.Header.Set("Content-Type", "application/json")
	// client := apiConfig.HttpClient
	// resp, err := client.Do(req)
	// if err != nil {
	// 	return fmt.Errorf("error sending request to worker. err: %v", err)
	// }
	// defer resp.Body.Close()
	// if resp.StatusCode != http.StatusOK {
	// 	return fmt.Errorf("worker returned non-200 status: %d", resp.StatusCode)
	// }
	// return nil

	// using google pub/sub
	ctx := context.Background()
	topic := apiConfig.PubSubClient.Topic("resume-analysis")
	result := topic.Publish(ctx, &pubsub.Message{
		Data: []byte("resume-analysis"),
		Attributes: map[string]string{
			"session_id": session.ID.String(),
		}})

	id, err := result.Get(ctx)
	if err != nil {
		// log.Println("Failed to publish message:", err)
		return err
	}
	log.Println("Published job ID:", session.ID.String(), "Message ID:", id)
	return nil
}
