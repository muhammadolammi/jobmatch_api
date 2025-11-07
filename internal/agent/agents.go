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
		APIKey: os.Getenv("GEMINI_API_KEY"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create model: %v", err)
	}

	customAgent, err := llmagent.New(llmagent.Config{
		Name:        "resume analyzer",
		Model:       model,
		Description: "Analyze Resume",
		Instruction: prompt(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create agent: %v", err)
	}

	return customAgent, err
}
