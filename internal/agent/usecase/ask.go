// Package usecase coordinates agent use cases through provider-neutral ports.
package usecase

import (
	"context"

	"github.com/insomnius/almira/internal/agent/entity"
)

// TextGenerator is the capability Ask needs. Each provider implements this port.
// JSON, HTTP, model identifiers, and credentials belong to the adapter.
type TextGenerator interface {
	Generate(context.Context, entity.Prompt) (string, error)
}

type Ask struct {
	generator TextGenerator
}

func NewAsk(generator TextGenerator) *Ask {
	return &Ask{generator: generator}
}

// Execute makes one model call. The agent loop will be a later increment.
func (a *Ask) Execute(ctx context.Context, text string) (string, error) {
	prompt, err := entity.NewPrompt(text)
	if err != nil {
		return "", err
	}
	return a.generator.Generate(ctx, prompt)
}
