// Package cmd provides Almira's command-line interface and wires its dependencies.
package cmd

import (
	"fmt"
	"os"

	"github.com/insomnius/almira/internal/agent/application"
	"github.com/insomnius/almira/internal/agent/infrastructure/openai"
	"github.com/spf13/cobra"
)

// NewRootCommand gives each invocation its own flags and dependencies.
func NewRootCommand() *cobra.Command {
	var prompt, model string
	root := &cobra.Command{
		Use:           "almira",
		Short:         "Send a prompt to Almira and print its answer",
		Example:       `  almira --prompt "Hello, Almira" --model gpt-4.1-mini`,
		Args:          cobra.NoArgs,
		SilenceUsage:  true,
		SilenceErrors: true, // main prints errors once and sets the exit status.
		RunE: func(command *cobra.Command, args []string) error {
			// Construct the provider only when running, so help needs no credentials.
			client, err := openai.NewClient(os.Getenv("OPENAI_API_KEY"), model, os.Getenv("OPENAI_BASE_URL"))
			if err != nil {
				return err
			}
			answer, err := application.NewAsk(client).Execute(command.Context(), prompt)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(command.OutOrStdout(), answer)
			return err
		},
	}
	root.Flags().StringVarP(&prompt, "prompt", "p", "", "text to send to Almira (required)")
	root.Flags().StringVarP(&model, "model", "m", os.Getenv("OPENAI_MODEL"), "model ID (defaults to OPENAI_MODEL)")
	// The flag was registered immediately above, so this annotation cannot fail.
	_ = root.MarkFlagRequired("prompt")
	return root
}
