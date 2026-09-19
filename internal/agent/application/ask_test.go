package application_test

import (
	"context"
	"errors"
	"testing"

	"github.com/insomnius/almira/internal/agent/application"
	"github.com/insomnius/almira/internal/agent/domain"
)

type generatorFunc func(context.Context, domain.Prompt) (string, error)

func (f generatorFunc) Generate(ctx context.Context, p domain.Prompt) (string, error) {
	return f(ctx, p)
}

func TestAskRejectsBlankInputBeforeCallingProvider(t *testing.T) {
	ask := application.NewAsk(generatorFunc(func(context.Context, domain.Prompt) (string, error) {
		t.Fatal("blank input reached the provider")
		return "", nil
	}))
	for _, input := range []string{"", " \n\t", "\u3000"} {
		if _, err := ask.Execute(context.Background(), input); !errors.Is(err, domain.ErrEmptyPrompt) {
			t.Errorf("input %q: got %v, want ErrEmptyPrompt", input, err)
		}
	}
}

func TestAskPreservesPromptAndContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	const input = "  explain this code:\n    hello()\n"
	ask := application.NewAsk(generatorFunc(func(got context.Context, p domain.Prompt) (string, error) {
		if got != ctx || p.Text() != input {
			t.Fatal("context or prompt changed at the application boundary")
		}
		return "answer", nil
	}))
	answer, err := ask.Execute(ctx, input)
	if err != nil || answer != "answer" {
		t.Fatalf("got %q, %v", answer, err)
	}
}
