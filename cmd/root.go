// Package cmd assembles Almira's command tree.
package cmd

import (
	"github.com/insomnius/almira/internal/agent"
	"github.com/spf13/cobra"
)

// NewRootCommand starts with the agent's single-prompt command.
func NewRootCommand() *cobra.Command {
	return agent.Provide()
}
