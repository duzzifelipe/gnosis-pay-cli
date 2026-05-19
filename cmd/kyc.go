package cmd

import (
	"fmt"

	"github.com/duzzifelipe/gnosis-pay/internal/apiclient"
	"github.com/duzzifelipe/gnosis-pay/internal/config"
	"github.com/spf13/cobra"
)

func init() {
	kycCmd.AddCommand(kycStartCmd)
	kycCmd.AddCommand(kycStatusCmd)
	rootCmd.AddCommand(kycCmd)
}

var kycCmd = &cobra.Command{
	Use:   "kyc",
	Short: "Manage KYC verification process",
	Long:  `Start KYC verification or check its status. KYC is completed via a Sumsub iframe in the browser.`,
}

var kycStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Get KYC verification URL",
	Long: `Retrieves the Sumsub iframe URL for KYC verification.
Open the URL in your browser to complete the KYC process.
After completing KYC, run 'gnosis-pay kyc status' to check the result.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _, err := config.GetAuthToken()
		if err != nil {
			return fmt.Errorf("get auth token: %w", err)
		}

		client := apiclient.New(config.DefaultAPIURL, token)

		// Get KYC integration URL
		fmt.Println("Requesting KYC verification URL...")
		resp, err := client.Get("/api/v1/kyc/integration")
		if err != nil {
			return fmt.Errorf("get KYC integration: %w", err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("get KYC integration failed (HTTP %d): %s", resp.StatusCode, resp.String())
		}

		var kycResp map[string]interface{}
		if err := resp.JSON(&kycResp); err != nil {
			return fmt.Errorf("parse KYC response: %w", err)
		}

		// Extract URL from response
		url := ""
		for _, key := range []string{"url", "link", "integrationUrl"} {
			if v, ok := kycResp[key].(string); ok {
				url = v
				break
			}
		}
		if url == "" {
			if data, ok := kycResp["data"].(map[string]interface{}); ok {
				for _, key := range []string{"url", "link", "integrationUrl"} {
					if v, ok := data[key].(string); ok {
						url = v
						break
					}
				}
			}
		}

		if url != "" {
			fmt.Println("\n=== KYC Verification URL ===")
			fmt.Println(url)
			fmt.Println("\nOpen this URL in your browser to complete KYC verification.")
			fmt.Println("After completing, run: gnosis-pay kyc status")
		} else {
			fmt.Println("\nFull response:")
			fmt.Println(apiclient.PrettyJSON(resp.Body))
		}

		return nil
	},
}

var kycStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Check KYC verification status",
	Long: `Queries the current KYC status for the authenticated user.
Possible statuses: notStarted, documentsRequested, pending, processing,
approved, resubmissionRequested, rejected, requiresAction`,
	RunE: func(cmd *cobra.Command, args []string) error {
		token, _, err := config.GetAuthToken()
		if err != nil {
			return fmt.Errorf("get auth token: %w", err)
		}

		client := apiclient.New(config.DefaultAPIURL, token)

		fmt.Println("Checking KYC status...")
		resp, err := client.Get("/api/v1/user")
		if err != nil {
			return fmt.Errorf("get user: %w", err)
		}
		if resp.StatusCode != 200 {
			return fmt.Errorf("get user failed (HTTP %d): %s", resp.StatusCode, resp.String())
		}

		var userResp map[string]interface{}
		if err := resp.JSON(&userResp); err != nil {
			return fmt.Errorf("parse user response: %w", err)
		}

		kycStatus, _ := userResp["kycStatus"].(string)
		isSOFAnswered, _ := userResp["isSourceOfFundsAnswered"].(bool)
		isPhoneValidated, _ := userResp["isPhoneValidated"].(bool)

		fmt.Printf("\n=== User Status ===\n")
		fmt.Printf("KYC Status:                %s\n", kycStatus)
		fmt.Printf("Source of Funds Answered:   %v\n", isSOFAnswered)
		fmt.Printf("Phone Validated:            %v\n", isPhoneValidated)

		return nil
	},
}
