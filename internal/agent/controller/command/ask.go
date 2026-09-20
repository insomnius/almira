// Package command translates CLI input into agent use-case calls.
package command

import (
	"fmt"

	"github.com/insomnius/almira/internal/agent/usecase"

	"github.com/spf13/cobra"
)

// NewAsk gives each invocation its own flags and lazily constructs the use case.
func NewAsk(newAsk func(model string) (*usecase.Ask, error), defaultModel string) *cobra.Command {
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
			ask, err := newAsk(model)
			if err != nil {
				return err
			}
			answer, err := ask.Execute(command.Context(), prompt)
			if err != nil {
				return err
			}
			_, err = fmt.Fprintln(command.OutOrStdout(), answer)
			return err
		},
	}
	root.Flags().StringVarP(&prompt, "prompt", "p", "", "text to send to Almira (required)")
	root.Flags().StringVarP(&model, "model", "m", defaultModel, "model ID (defaults to OPENAI_MODEL)")
	// The flag was registered immediately above, so this annotation cannot fail.
	_ = root.MarkFlagRequired("prompt")
	return root
}
