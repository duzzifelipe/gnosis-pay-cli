package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/duzzifelipe/gnosis-pay/internal/apiclient"
	"github.com/duzzifelipe/gnosis-pay/internal/config"
	"github.com/duzzifelipe/gnosis-pay/internal/signer"
	"github.com/spf13/cobra"
)

var authTTL int

func init() {
	authCmd.Flags().IntVar(&authTTL, "ttl", 36000, "JWT time-to-live in seconds (min 3600, max 86400)")
	rootCmd.AddCommand(authCmd)
}

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate with Gnosis Pay using SIWE",
	Long: `Performs Sign-In with Ethereum (SIWE) authentication.

Required environment variables:
  GNOSIS_PAY_PRIVATE_KEY    Hex-encoded Ethereum private key

Optional environment variables (default to localhost):
  GNOSIS_PAY_DOMAIN         Domain for SIWE message (e.g., yourapp.com)
  GNOSIS_PAY_URI            URI for SIWE message (e.g., https://yourapp.com)

The domain/URI must NOT be localhost or 127.0.0.1.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		domain := config.Domain()
		uri := config.URI()
		fmt.Printf("Using SIWE domain: %s\n", domain)
		fmt.Printf("Using SIWE URI: %s\n", uri)
		// Load private key
		hexKey, err := config.PrivateKey()
		if err != nil {
			return err
		}

		s, err := signer.New(hexKey)
		if err != nil {
			return err
		}

		address := s.Address().Hex()
		fmt.Printf("Wallet address: %s\n", address)

		client := apiclient.NewWithOrigin(config.DefaultAPIURL, "", uri)

		// Step 1: Get nonce
		fmt.Println("Requesting nonce...")
		resp, err := client.Get("/api/v1/auth/nonce")
		if err != nil {
			return fmt.Errorf("get nonce: %w", err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("get nonce failed (HTTP %d): %s", resp.StatusCode, resp.String())
		}

		var nonceResp struct {
			Nonce string `json:"nonce"`
		}
		if err := resp.JSON(&nonceResp); err != nil {
			// Some APIs return the nonce as a plain string in a data field
			var dataResp struct {
				Data string `json:"data"`
			}
			if err2 := resp.JSON(&dataResp); err2 != nil {
				// Try as plain text nonce wrapper
				nonceResp.Nonce = string(resp.Body)
			} else {
				nonceResp.Nonce = dataResp.Data
			}
		}
		if nonceResp.Nonce == "" {
			// Try to extract from a nested response
			var raw map[string]interface{}
			if err := resp.JSON(&raw); err == nil {
				if n, ok := raw["nonce"].(string); ok {
					nonceResp.Nonce = n
				} else if d, ok := raw["data"]; ok {
					if dm, ok := d.(map[string]interface{}); ok {
						if n, ok := dm["nonce"].(string); ok {
							nonceResp.Nonce = n
						}
					} else if ds, ok := d.(string); ok {
						nonceResp.Nonce = ds
					}
				}
			}
		}
		nonceResp.Nonce = strings.TrimSpace(nonceResp.Nonce)
		if nonceResp.Nonce == "" {
			return fmt.Errorf("could not parse nonce from response: %s", resp.String())
		}

		fmt.Printf("Nonce: %s\n", nonceResp.Nonce)

		// Step 2: Build and sign SIWE message
		message := signer.BuildSIWEMessage(
			domain,
			uri,
			address,
			nonceResp.Nonce,
			config.GnosisChainID,
		)

		fmt.Printf("SIWE message:\n---\n%s\n---\n", message)
		fmt.Println("Signing SIWE message...")
		signature, err := s.SignMessage(message)
		if err != nil {
			return fmt.Errorf("sign SIWE message: %w", err)
		}

		// Step 3: Submit challenge
		fmt.Println("Submitting authentication challenge...")
		challengeBody := map[string]interface{}{
			"message":      message,
			"signature":    signature,
			"ttlInSeconds": authTTL,
		}

		resp, err = client.Post("/api/v1/auth/challenge", challengeBody)
		if err != nil {
			return fmt.Errorf("submit challenge: %w", err)
		}
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			fmt.Printf("Challenge response: %s\n", resp.String())
			for name, values := range resp.Headers {
				for _, value := range values {
					fmt.Printf(" -> %s: %s\n", name, value)
				}
			}
			return fmt.Errorf("auth challenge failed (HTTP %d): %s", resp.StatusCode, resp.String())
		}

		// Extract JWT from response
		var authResp map[string]interface{}
		if err := resp.JSON(&authResp); err != nil {
			return fmt.Errorf("parse auth response: %w", err)
		}

		jwt := extractString(authResp, "token", "jwt", "accessToken", "access_token")
		if jwt == "" {
			// Try nested under "data"
			if data, ok := authResp["data"].(map[string]interface{}); ok {
				jwt = extractString(data, "token", "jwt", "accessToken", "access_token")
			}
		}
		if jwt == "" {
			fmt.Fprintf(os.Stderr, "Warning: could not extract JWT automatically. Full response:\n%s\n",
				apiclient.PrettyJSON(resp.Body))
			return fmt.Errorf("could not extract JWT from response")
		}

		// Save state
		state, err := config.LoadState()
		if err != nil {
			return fmt.Errorf("load state: %w", err)
		}
		state.JWT = jwt
		state.Address = address
		if err := state.Save(); err != nil {
			return fmt.Errorf("save state: %w", err)
		}

		fmt.Println("Authentication successful! JWT saved to state.")
		fmt.Printf("JWT (first 50 chars): %s...\n", jwt[:min(50, len(jwt))])
		return nil
	},
}

// extractString tries multiple keys and returns the first string found.
func extractString(m map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := m[k].(string); ok && v != "" {
			return v
		}
	}
	return ""
}
