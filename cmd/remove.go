package cmd

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var removeCmd = &cobra.Command{
	Use:   "remove [account-name]",
	Short: "Remove a GitHub account configuration",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		accountName := args[0]

		// Load config
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		// Check if account exists
		account := config.getAccount(accountName)
		if account == nil {
			availableAccounts := make([]string, len(config.Accounts))
			for i, acc := range config.Accounts {
				availableAccounts[i] = acc.Name
			}
			return fmt.Errorf("❌ Account '%s' not found. Available accounts: %s", 
				accountName, strings.Join(availableAccounts, ", "))
		}

		// Confirm removal
		reader := bufio.NewReader(os.Stdin)
		fmt.Printf("⚠️  Are you sure you want to remove account '%s'? [y/N]: ", accountName)
		resp, _ := reader.ReadString('\n')
		resp = strings.ToLower(strings.TrimSpace(resp))

		if resp != "y" && resp != "yes" {
			fmt.Println("❌ Account removal cancelled")
			return nil
		}

		// Remove from accounts list
		var newAccounts []Account
		for _, acc := range config.Accounts {
			if acc.Name != accountName {
				newAccounts = append(newAccounts, acc)
			}
		}
		config.Accounts = newAccounts

		// Update current account if needed
		if config.CurrentAccount == accountName {
			if len(config.Accounts) > 0 {
				config.CurrentAccount = config.Accounts[0].Name
				fmt.Printf("🔄 Current account switched to '%s'\n", config.CurrentAccount)
			} else {
				config.CurrentAccount = ""
				fmt.Println("📝 No accounts remaining")
			}
		}

		// Save config
		if err := config.saveConfig(); err != nil {
			return fmt.Errorf("failed to save config: %w", err)
		}

		fmt.Printf("✅ Account '%s' removed successfully\n", accountName)
		
		// Optionally remove SSH key
		if account.SSHKey != "" {
			fmt.Printf("\n💡 SSH key still exists at: %s\n", account.SSHKey)
			fmt.Print("🗑️  Do you want to remove the SSH key files? [y/N]: ")
			resp, _ := reader.ReadString('\n')
			resp = strings.ToLower(strings.TrimSpace(resp))

			if resp == "y" || resp == "yes" {
				// Remove private key
				if err := os.Remove(account.SSHKey); err != nil {
					fmt.Printf("⚠️  Could not remove private key: %v\n", err)
				} else {
					fmt.Printf("🗑️  Removed: %s\n", account.SSHKey)
				}

				// Remove public key
				pubKeyPath := account.SSHKey + ".pub"
				if err := os.Remove(pubKeyPath); err != nil {
					fmt.Printf("⚠️  Could not remove public key: %v\n", err)
				} else {
					fmt.Printf("🗑️  Removed: %s\n", pubKeyPath)
				}
			}
		}

		fmt.Println("\n💡 Note: You may want to:")
		fmt.Printf("   - Remove the SSH key from GitHub: https://github.com/settings/ssh\n")
		fmt.Printf("   - Clean up any conditional includes in ~/.gitconfig manually\n")

		return nil
	},
}

func init() {
	RootCmd.AddCommand(removeCmd)
}
