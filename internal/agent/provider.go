// Package agent wires the agent module's controllers, use cases, and adapters.
package agent

import (
	"os"

	"github.com/insomnius/almira/internal/agent/controller/command"
	"github.com/insomnius/almira/internal/agent/usecase"
	"github.com/insomnius/almira/internal/provider/openai"
	"github.com/spf13/cobra"
)

// Provide assembles the CLI. Provider construction is deferred until execution
// so displaying help never needs credentials or opens a provider connection.
func Provide() *cobra.Command {
	return command.NewAsk(func(model string) (*usecase.Ask, error) {
		client, err := openai.NewClient(os.Getenv("OPENAI_API_KEY"), model, os.Getenv("OPENAI_BASE_URL"))
		if err != nil {
			return nil, err
		}
		return usecase.NewAsk(client), nil
	}, os.Getenv("OPENAI_MODEL"))
}
