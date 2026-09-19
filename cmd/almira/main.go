package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"

	"github.com/insomnius/almira/internal/agent/application"
	"github.com/insomnius/almira/internal/agent/infrastructure/openai"
)

func main() {
	prompt := flag.String("prompt", "", "text to send to Almira (required)")
	model := flag.String("model", os.Getenv("OPENAI_MODEL"), "model ID (defaults to OPENAI_MODEL)")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "almira: use -prompt to provide your message")
		os.Exit(1)
	}

	// This is the composition root: only the entry point chooses an adapter.
	client, err := openai.NewClient(os.Getenv("OPENAI_API_KEY"), *model, os.Getenv("OPENAI_BASE_URL"))
	if err != nil {
		fail(err)
	}
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	answer, err := application.NewAsk(client).Execute(ctx, *prompt)
	if err != nil {
		fail(err)
	}
	fmt.Println(answer)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "almira:", err)
	os.Exit(1)
}
