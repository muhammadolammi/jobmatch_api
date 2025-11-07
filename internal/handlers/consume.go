package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	_ "github.com/lib/pq"
	customagent "github.com/muhammadolammi/jobmatchapi/internal/agent"
	"github.com/muhammadolammi/jobmatchapi/internal/database"
	"github.com/muhammadolammi/jobmatchapi/internal/helpers"
	"github.com/streadway/amqp"
	"google.golang.org/adk/agent"
	"google.golang.org/adk/runner"
	"google.golang.org/adk/session"
	"google.golang.org/genai"
)

func aggregateResult(results *AnalysesResults, resultStr string, hasError bool, errorMsg string) {
	result := AnalysesResult{}

	switch {
	case hasError:
		result.IsErrorResult = true
		result.Error = errorMsg

	case resultStr == "":
		result.IsErrorResult = true
		result.Error = "empty response from agent"

	default:
		if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
			result.IsErrorResult = true
			result.Error = "json unmarshal error: " + err.Error()
		}
	}

	results.Results = append(results.Results, result)
}

func callAgent(currentSession Session, apiConfig *Config) error {
	ctx := context.Background()

	analyzer, err := customagent.GetAgent()
	if err != nil {
		return fmt.Errorf("failed to create agent: %w", err)
	}

	r, err := runner.New(runner.Config{
		AppName:        analyzer.Name(),
		Agent:          analyzer,
		SessionService: session.InMemoryService(),
	})
	if err != nil {
		return fmt.Errorf("failed to create runner: %w", err)
	}
	resumes, err := apiConfig.DB.GetResumesBySession(context.Background(), currentSession.ID)
	if err != nil {
		return fmt.Errorf("error getting resumes for session: %v from db. err: %v", currentSession.ID, err)
	}
	results := &AnalysesResults{
		SessionID: currentSession.ID,
	}
	for _, resume := range resumes {
		awsclient := s3.NewFromConfig(*apiConfig.AwsConfig, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(fmt.Sprintf("https://%s.r2.cloudflarestorage.com", apiConfig.R2.AccountID))
		})
		fileBytes, err := helpers.DownloadFromR2(ctx, awsclient, apiConfig.R2.Bucket, resume.ObjectKey)

		if err != nil {
			log.Printf("⚠️ failed to download %s: %v", resume.ObjectKey, err)
			aggregateResult(results, "", true, "file download error: ⚠️ failed to download "+resume.ObjectKey+"err: "+err.Error())
			continue
		}
		resumeText, err := helpers.ExtractResumeText(resume.Mime, fileBytes)
		if err != nil {
			log.Printf("⚠️ failed to extract text for %s. err: %v. currentsession: %v\n", resume.StorageUrl, err, currentSession.ID)
			aggregateResult(results, "", true, "text extraction error: "+err.Error())
			continue
		}

		msg := fmt.Sprintf(
			"Job Title:\n%s\n\nJob Description:\n%s\n\nResume:\n%s",
			currentSession.JobTitle,
			currentSession.JobDescription,
			resumeText,
		)

		stream := r.Run(ctx, currentSession.UserID.String(), currentSession.ID.String(), &genai.Content{
			Role: "user",
			Parts: []*genai.Part{
				{Text: msg},
			},
		}, agent.RunConfig{})

		var finalOutput string
		var streamError error

		for event, err := range stream {
			if err != nil {
				streamError = err
				break
			}
			if event != nil && event.IsFinalResponse() && len(event.Content.Parts) > 0 {
				finalOutput = event.Content.Parts[0].Text
			}
		}
		if streamError != nil {
			aggregateResult(results, "", true, "agent stream error: "+streamError.Error())
		} else {
			aggregateResult(results, finalOutput, false, "")
		}

	}

	resultsJSON, err := json.Marshal(results.Results)
	if err != nil {
		return fmt.Errorf("failed to marshal analyses results: %w", err)
	}
	err = apiConfig.DB.CreateOrUpdateAnalysesResults(context.Background(), database.CreateOrUpdateAnalysesResultsParams{
		Results:   resultsJSON,
		SessionID: results.SessionID,
	})
	if err != nil {
		return fmt.Errorf("failed to save agent result: %w", err)
	}

	return nil
}
func worker(id int, apiConfig *Config, wg *sync.WaitGroup) {
	defer wg.Done()
	//    to consume message on the queue
	conn, err := amqp.Dial(apiConfig.RABBITMQUrl)
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

		err = callAgent(session, apiConfig)
		if err != nil {
			log.Printf("error running agent for session_id: %v. err: %v", session.ID, err)
			continue
		}
		// update session status
		err = apiConfig.DB.UpdateSessionStatus(context.Background(), database.UpdateSessionStatusParams{
			Status: "completed",
			ID:     session.ID,
		})
		if err != nil {
			log.Printf("error updating session status in db to completed for  session_id: %v. err: %v", session.ID, err)
			continue
		}
	}

}

func (apiConfig *Config) StartConsumerWorkerPool(numWorkers int) {
	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := range numWorkers {
		log.Println("worker id ", i+1, "started")
		// wg.Done()
		// continue
		go worker(i, apiConfig, &wg)
	}
	wg.Wait() // block until all workers finish

}
