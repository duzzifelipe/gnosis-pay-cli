package cmd

import (
	"fmt"

	"github.com/duzzifelipe/gnosis-pay/internal/apiclient"
	"github.com/duzzifelipe/gnosis-pay/internal/config"
	"github.com/spf13/cobra"
)

var signupEmail string
var signupPartnerID string
var signupCoupon string

func init() {
	signupCmd.Flags().StringVar(&signupEmail, "email", "", "User email address (required)")
	signupCmd.Flags().StringVar(&signupPartnerID, "partner-id", "", "Partner ID (optional for permissionless)")
	signupCmd.Flags().StringVar(&signupCoupon, "coupon", "", "Referral coupon code (optional)")
	_ = signupCmd.MarkFlagRequired("email")
	rootCmd.AddCommand(signupCmd)
}

var signupCmd = &cobra.Command{
	Use:   "signup",
	Short: "Register a new user with Gnosis Pay",
	Long: `Registers the authenticated wallet as a new Gnosis Pay user.
Requires prior authentication (run 'gnosis-pay auth' first).`,
	RunE: func(cmd *cobra.Command, args []string) error {
		state, err := config.LoadState()
		if err != nil {
			return fmt.Errorf("load state: %w", err)
		}
		// if state.JWT == "" {
		// 	return fmt.Errorf("not authenticated. Run 'gnosis-pay auth' first")
		// }

		client := apiclient.New(config.DefaultAPIURL, state.JWT)

		// Check if user is already registered
		// fmt.Println("Checking if user is already registered...")
		// resp, err := client.Get("/api/v1/user")
		// if err != nil {
		// 	return fmt.Errorf("check user: %w", err)
		// }
		// if resp.StatusCode == 200 {
		// 	fmt.Println("User is already registered!")
		// 	fmt.Println(apiclient.PrettyJSON(resp.Body))

		// 	// Extract user ID if present
		// 	var userResp map[string]interface{}
		// 	if err := resp.JSON(&userResp); err == nil {
		// 		if id := extractNestedString(userResp, "id"); id != "" {
		// 			state.UserID = id
		// 			_ = state.Save()
		// 		}
		// 	}
		// 	return nil
		// }

		// Register new user
		fmt.Printf("Registering user with email: %s\n", signupEmail)
		body := map[string]interface{}{
			"authEmail": signupEmail,
		}
		if signupPartnerID != "" {
			body["partnerId"] = signupPartnerID
		}
		if signupCoupon != "" {
			body["referralCouponCode"] = signupCoupon
		}

		resp, err := client.Post("/api/v1/auth/signup", body)
		if err != nil {
			return fmt.Errorf("signup: %w", err)
		}
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return fmt.Errorf("signup failed (HTTP %d): %s", resp.StatusCode, resp.String())
		}

		fmt.Println("User registered successfully!")
		fmt.Println(apiclient.PrettyJSON(resp.Body))

		// Save email to state
		state.Email = signupEmail
		if err := state.Save(); err != nil {
			return fmt.Errorf("save state: %w", err)
		}

		return nil
	},
}

// extractNestedString tries to get a string from top-level or nested under "data".
// func extractNestedString(m map[string]interface{}, key string) string {
// 	if v, ok := m[key].(string); ok {
// 		return v
// 	}
// 	if data, ok := m["data"].(map[string]interface{}); ok {
// 		if v, ok := data[key].(string); ok {
// 			return v
// 		}
// 	}
// 	return ""
// }
