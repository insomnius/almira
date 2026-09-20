package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/insomnius/almira/cmd"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "almira:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	return cmd.NewRootCommand().ExecuteContext(ctx)
}
