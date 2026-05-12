package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gnosis-pay",
	Short: "Gnosis Pay CLI - Permissionless integration",
	Long: `A CLI tool implementing the full Gnosis Pay permissionless integration flow.

Workflow (run each step in order):

  1. gnosis-pay auth                     Authenticate with SIWE

Required environment variables:
  GNOSIS_PAY_PRIVATE_KEY   Hex-encoded Ethereum private key

Optional environment variables (for SIWE - must not be localhost):
  GNOSIS_PAY_DOMAIN        Domain for SIWE message (default: example.com)
  GNOSIS_PAY_URI           URI for SIWE message (default: https://example.com)

State is persisted to .gnosis-pay-state.json in the working directory.`,
}

// Execute runs the root command.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
