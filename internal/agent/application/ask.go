// Package application coordinates agent use cases through provider-neutral ports.
package application

import (
	"context"

	"github.com/insomnius/almira/internal/agent/domain"
)

// TextGenerator is the capability Ask needs. Each provider implements this port.
// JSON, HTTP, model identifiers, and credentials belong to the adapter.
type TextGenerator interface {
	Generate(context.Context, domain.Prompt) (string, error)
}

type Ask struct {
	generator TextGenerator
}

func NewAsk(generator TextGenerator) *Ask {
	return &Ask{generator: generator}
}

// Execute makes one model call. The agent loop will be a later increment.
func (a *Ask) Execute(ctx context.Context, text string) (string, error) {
	prompt, err := domain.NewPrompt(text)
	if err != nil {
		return "", err
	}
	return a.generator.Generate(ctx, prompt)
}
