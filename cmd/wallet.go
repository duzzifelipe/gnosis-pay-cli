package cmd

import (
	"fmt"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(walletCmd)
	walletCmd.AddCommand(walletGenerateCmd)
}

var walletCmd = &cobra.Command{
	Use:   "wallet",
	Short: "Wallet utilities",
}

var walletGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate a new Ethereum wallet",
	RunE: func(cmd *cobra.Command, args []string) error {
		key, err := crypto.GenerateKey()
		if err != nil {
			return fmt.Errorf("generate key: %w", err)
		}

		privateKeyHex := fmt.Sprintf("%x", key.D.Bytes())
		address := crypto.PubkeyToAddress(key.PublicKey).Hex()

		fmt.Printf("Address:     %s\n", address)
		fmt.Printf("Private key: %s\n", privateKeyHex)
		fmt.Println()
		fmt.Printf("Export and retry auth:\n")
		fmt.Printf("  export GNOSIS_PAY_PRIVATE_KEY=%s\n", privateKeyHex)

		return nil
	},
}
