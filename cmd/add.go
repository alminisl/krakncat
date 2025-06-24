package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var addCmd = &cobra.Command{
	Use:   "add",
	Short: "Add a new GitHub account",
	Long:  "Add a new GitHub account with SSH key configuration",
	RunE: func(cmd *cobra.Command, args []string) error {
		reader := bufio.NewReader(os.Stdin)

		// Get account name
		fmt.Print("💬 Account name (e.g., 'work', 'personal'): ")
		name, _ := reader.ReadString('\n')
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("account name cannot be empty")
		}

		// Get email
		fmt.Print("📧 Email address: ")
		email, _ := reader.ReadString('\n')
		email = strings.TrimSpace(email)
		if email == "" {
			return fmt.Errorf("email cannot be empty")
		}

		// Get GitHub username
		fmt.Print("👤 GitHub username: ")
		username, _ := reader.ReadString('\n')
		username = strings.TrimSpace(username)
		if username == "" {
			return fmt.Errorf("GitHub username cannot be empty")
		}

		// Check for existing SSH key
		homeDir, _ := os.UserHomeDir()
		defaultSSHKey := filepath.Join(homeDir, ".ssh", fmt.Sprintf("id_ed25519_gh_%s", name))
		
		fmt.Printf("🔑 SSH key path [%s]: ", defaultSSHKey)
		sshKeyInput, _ := reader.ReadString('\n')
		sshKeyInput = strings.TrimSpace(sshKeyInput)
		
		sshKey := defaultSSHKey
		if sshKeyInput != "" {
			sshKey = sshKeyInput
		}

		// Verify SSH key exists
		if _, err := os.Stat(sshKey); os.IsNotExist(err) {
			fmt.Printf("⚠️  SSH key not found at %s\n", sshKey)
			fmt.Print("🤔 Do you want to generate it now? [Y/n]: ")
			resp, _ := reader.ReadString('\n')
			resp = strings.ToLower(strings.TrimSpace(resp))
			
			if resp == "y" || resp == "" {
				// Generate SSH key
				if err := generateSSHKey(name, email, sshKey); err != nil {
					return fmt.Errorf("failed to generate SSH key: %w", err)
				}
			} else {
				return fmt.Errorf("cannot add account without SSH key")
			}
		}

		// Load config and add account
		config, err := loadConfig()
		if err != nil {
			return fmt.Errorf("failed to load config: %w", err)
		}

		account := Account{
			Name:     name,
			Email:    email,
			SSHKey:   sshKey,
			Username: username,
		}

		if err := config.addAccount(account); err != nil {
			return fmt.Errorf("failed to add account: %w", err)
		}

		fmt.Printf("✅ Account '%s' added successfully!\n", name)
		fmt.Printf("🔗 SSH Host: github.com-%s\n", name)
		fmt.Printf("📂 Config saved to: %s\n", getConfigPath())

		return nil
	},
}

func init() {
	RootCmd.AddCommand(addCmd)
}
