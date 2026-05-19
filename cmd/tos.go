package cmd

import (
	"fmt"

	"github.com/duzzifelipe/gnosis-pay/internal/apiclient"
	"github.com/duzzifelipe/gnosis-pay/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(tosCmd)
}

var tosCmd = &cobra.Command{
	Use:   "tos",
	Short: "Accept all pending Terms of Service",
	Long: `Fetches all terms of service and accepts any that haven't been accepted yet.
Requires prior authentication and registration.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _, err := config.GetAuthToken()
		if err != nil {
			return fmt.Errorf("get auth token: %w", err)
		}

		client := apiclient.New(config.DefaultAPIURL, token)

		// Fetch terms
		fmt.Println("Fetching terms of service...")
		resp, err := client.Get("/api/v1/user/terms")
		if err != nil {
			return fmt.Errorf("get terms: %w", err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("get terms failed (HTTP %d): %s", resp.StatusCode, resp.String())
		}

		// Parse terms response - try multiple formats
		type TermType struct {
			Type    string `json:"type,omitempty"`
			Accepted bool   `json:"accepted,omitempty"`
			Version string `json:"currentVersion,omitempty"`
			URL     string `json:"url,omitempty"`
		}

		var result struct {
			Terms []TermType `json:"terms"`
		}
		
		if err := resp.JSON(&result); err != nil {
			fmt.Printf("Response does not match expected format: %v\n", err)
			return nil;
		}

		if len(result.Terms) == 0 {
			fmt.Println("No terms of service found or all already accepted.")
			return nil
		}

		acceptedCount := 0
		for _, term := range result.Terms {
			if term.Accepted {
				fmt.Printf("%s [already accepted] (%s)\n", term.Type, term.Version)
				continue
			}

			fmt.Printf("%s [accepting] (%s)\n", term.Type, term.Version)
			if term.URL != "" {
				fmt.Printf(" -> URL: %s\n", term.URL)
			}

			// Accept the term
			body := map[string]interface{}{
				"terms":   term.Type,
				"version": term.Version,
			}

			acceptResp, err := client.Post("/api/v1/user/terms", body)
			if err != nil {
				return fmt.Errorf(" -> accept term %s: %w", term.Type, err)
			}
			if acceptResp.StatusCode != 200 && acceptResp.StatusCode != 201 {
				fmt.Printf(" -> Warning: failed to accept (HTTP %d): %s\n",
					acceptResp.StatusCode, acceptResp.String())
				continue
			}

			fmt.Printf(" -> Accepted!\n")
			acceptedCount++
		}

		if acceptedCount > 0 {
			fmt.Printf("\nAccepted %d terms of service.\n", acceptedCount)
		} else {
			fmt.Println("\nAll terms were already accepted.")
		}

		return nil
	},
}
