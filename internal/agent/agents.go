package agent

import (
	"context"

	"github.com/go-a2a/adk-go/agent"
	"github.com/go-a2a/adk-go/model"
)

func GetAgent() (*agent.LLMAgent, error) {
	ctx := context.Background()

	m, err := model.NewGoogleModel("gemini-2.0-pro")
	if err != nil {
		return *agent.LLMAgent{}, err
	}
	defer m.Close()
	customAgent, err := agent.NewLLMAgent(ctx, "analyzer",
		agent.WithModel(m),
		agent.WithInstruction(prompt()),
	)

	return customAgent, err
}
