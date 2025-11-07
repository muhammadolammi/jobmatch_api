package customagent

import (
	"context"
	"fmt"
	"os"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/genai"
)

func GetAgent() (agent.Agent, error) {
	ctx := context.Background()

	model, err := gemini.NewModel(ctx, "gemini-2.5-flash", &genai.ClientConfig{
		APIKey: os.Getenv("GOOGLE_API_KEY"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create model: %v", err)
	}

	customAgent, err := llmagent.New(llmagent.Config{
		Name:        "resume analyzer",
		Model:       model,
		Description: "Analyze Resume",
		// GlobalInstruction: prompt(),
		// Instruction:       input,
		Instruction: prompt(),

		// Tools: []tool.Tool{
		// 	geminitool.GoogleSearch{},
		// },
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %v", err)
	}

	return customAgent, err
}

func AnalyzeResume(apiKey, input string) (string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, apiKey)
	if err != nil {
		return "", fmt.Errorf("failed to create genai client: %v", err)
	}
	defer client.Close()

	model := client.GenerativeModel("gemini-1.5-flash") // or 1.5-pro if available

	resp, err := model.GenerateContent(ctx, genai.Text(input))
	if err != nil {
		return "", fmt.Errorf("model error: %v", err)
	}

	if len(resp.Candidates) > 0 {
		return resp.Candidates[0].Content.Parts[0].(genai.Text).Text(), nil
	}
	return "", fmt.Errorf("no response from model")
}
